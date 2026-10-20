#!/usr/bin/env python3
"""Small room messenger API with encrypted messages and local accounts."""
from http.cookies import CookieError, SimpleCookie
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse
from contextlib import contextmanager
import base64
import hashlib
import hmac
import ipaddress
import json
import os
import secrets
import sqlite3
import threading
import time

from cryptography.exceptions import InvalidTag
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

ROOT = Path(__file__).parent
DATA_DIR = Path(os.environ.get("MESSENGER_DATA_DIR", str(ROOT)))
DB = DATA_DIR / "messages.sqlite3"
KEY_FILE = Path(os.environ.get("MESSENGER_KEY_FILE", "/etc/messenger/messenger.key"))
LOCK = threading.RLock()
RATE = {}
STALE_SECONDS = 90
SESSION_SECONDS = 30 * 24 * 60 * 60
PASSWORD_ITERATIONS = 310_000
MAX_NICK, MAX_ROOM_NAME, MAX_MESSAGE = 32, 64, 2000
ALLOWED_ORIGINS = {"https://max.smlt.lol", "https://chat.smlt.lol"}
AVATARS = {
    **{f"avatar{i}": f"avatar{i}" for i in range(1, 21)},
    **{name: "legacy" for name in ("moon", "sun", "star", "fox", "cat", "dog", "frog", "owl", "ghost", "robot", "coffee", "leaf")},
}
COLORS = {"violet": "#b9a7ff", "blue": "#76b9ff", "green": "#75d6a2", "yellow": "#f3ca68", "orange": "#f3a16b", "pink": "#f39ac4", "cyan": "#6edbd6", "red": "#ff8d9b"}
DEFAULT_AVATAR, DEFAULT_COLOR = "avatar1", "violet"
KEY = None


def now_ts():
    return int(time.time())


def token_hash(token):
    return hashlib.sha256(token.encode("utf-8")).hexdigest() if token else ""


def clean_text(value, limit):
    value = str(value or "").replace("\x00", "").strip()
    return "".join(ch for ch in value if ord(ch) >= 32 or ch in "\n\t")[:limit]


def nick_value(value):
    return clean_text(value, MAX_NICK)


def nick_key(value):
    return " ".join(nick_value(value).split()).casefold()


def valid_profile(avatar, color):
    return avatar in AVATARS and color in COLORS


def password_hash(password, salt):
    return hashlib.pbkdf2_hmac("sha256", password.encode("utf-8"), salt, PASSWORD_ITERATIONS)


def valid_password(password):
    return isinstance(password, str) and 12 <= len(password) <= 128


def new_recovery_code():
    raw = secrets.token_hex(16).upper()
    return "-".join(raw[i:i + 4] for i in range(0, len(raw), 4))


def recovery_value(code):
    return "".join(ch for ch in str(code or "").upper() if ch.isalnum())


def key_bytes():
    if not KEY_FILE.exists():
        raise RuntimeError(f"encryption key is missing: {KEY_FILE}")
    raw = base64.urlsafe_b64decode(KEY_FILE.read_text().strip().encode())
    if len(raw) != 32:
        raise RuntimeError("encryption key must be 32 bytes")
    return raw


def encrypt_body(text):
    nonce = secrets.token_bytes(12)
    encrypted = AESGCM(KEY).encrypt(nonce, text.encode("utf-8"), None)
    return "enc1:" + base64.urlsafe_b64encode(nonce + encrypted).decode("ascii")


def decrypt_body(value):
    if not value.startswith("enc1:"):
        return value
    try:
        raw = base64.urlsafe_b64decode(value[5:].encode("ascii"))
        return AESGCM(KEY).decrypt(raw[:12], raw[12:], None).decode("utf-8")
    except (ValueError, InvalidTag, UnicodeDecodeError):
        return "[сообщение недоступно]"


def db_init():
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    with db_connection() as con:
        con.execute("PRAGMA busy_timeout=5000")
        con.execute("PRAGMA journal_mode=WAL")
        con.execute("CREATE TABLE IF NOT EXISTS rooms (code TEXT PRIMARY KEY, owner_token TEXT NOT NULL, name TEXT NOT NULL DEFAULT 'Новая комната', created_at TEXT NOT NULL DEFAULT (datetime('now')))")
        con.execute("""CREATE TABLE IF NOT EXISTS messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT, room TEXT NOT NULL, nickname TEXT NOT NULL,
            body TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT (datetime('now')),
            author_token TEXT NOT NULL DEFAULT '', avatar_id TEXT NOT NULL DEFAULT 'moon',
            color_id TEXT NOT NULL DEFAULT 'violet', body_enc INTEGER NOT NULL DEFAULT 0)""")
        for col, default in (("avatar_id", "moon"), ("color_id", "violet"), ("body_enc", "0")):
            if col not in {r[1] for r in con.execute("PRAGMA table_info(messages)")}:
                con.execute(f"ALTER TABLE messages ADD COLUMN {col} TEXT NOT NULL DEFAULT '{default}'")
        con.execute("""CREATE TABLE IF NOT EXISTS participants (
            room TEXT NOT NULL, token_hash TEXT NOT NULL, nickname TEXT NOT NULL,
            nickname_key TEXT NOT NULL, last_seen INTEGER NOT NULL,
            avatar_id TEXT NOT NULL DEFAULT 'moon', color_id TEXT NOT NULL DEFAULT 'violet',
            PRIMARY KEY(room, token_hash))""")
        for col, default in (("avatar_id", "moon"), ("color_id", "violet")):
            if col not in {r[1] for r in con.execute("PRAGMA table_info(participants)")}:
                con.execute(f"ALTER TABLE participants ADD COLUMN {col} TEXT NOT NULL DEFAULT '{default}'")
        con.execute("""CREATE TABLE IF NOT EXISTS accounts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            nickname TEXT NOT NULL, registered_nickname TEXT NOT NULL DEFAULT '', nickname_key TEXT NOT NULL UNIQUE,
            password_salt BLOB NOT NULL, password_hash BLOB NOT NULL,
            recovery_salt BLOB NOT NULL, recovery_hash BLOB NOT NULL,
            token_hash TEXT NOT NULL UNIQUE,
            avatar_id TEXT NOT NULL DEFAULT 'avatar1', color_id TEXT NOT NULL DEFAULT 'violet',
            created_at INTEGER NOT NULL)""")
        if "registered_nickname" not in {r[1] for r in con.execute("PRAGMA table_info(accounts)")}:
            con.execute("ALTER TABLE accounts ADD COLUMN registered_nickname TEXT NOT NULL DEFAULT ''")
        con.execute("UPDATE accounts SET registered_nickname=nickname WHERE registered_nickname=''")
        con.execute("""CREATE TABLE IF NOT EXISTS account_sessions (
            token_hash TEXT PRIMARY KEY, account_id INTEGER NOT NULL,
            expires_at INTEGER NOT NULL, created_at INTEGER NOT NULL,
            FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE)""")
        con.execute("CREATE INDEX IF NOT EXISTS idx_account_sessions_expiry ON account_sessions(expires_at)")
        con.execute("""CREATE TABLE IF NOT EXISTS registration_ips (
            ip_hash TEXT NOT NULL, account_id INTEGER NOT NULL UNIQUE, created_at INTEGER NOT NULL,
            FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE)""")
        con.execute("CREATE INDEX IF NOT EXISTS idx_registration_ips_hash ON registration_ips(ip_hash)")
        con.execute("""CREATE TABLE IF NOT EXISTS account_rooms (
            account_id INTEGER NOT NULL, room TEXT NOT NULL, added_at INTEGER NOT NULL,
            PRIMARY KEY(account_id, room), FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE)""")
        con.execute("""CREATE TABLE IF NOT EXISTS friendships (
            account_a INTEGER NOT NULL, account_b INTEGER NOT NULL, created_at INTEGER NOT NULL,
            PRIMARY KEY(account_a, account_b), CHECK(account_a < account_b))""")
        con.execute("""CREATE TABLE IF NOT EXISTS friend_requests (
            id INTEGER PRIMARY KEY AUTOINCREMENT, requester_id INTEGER NOT NULL,
            recipient_id INTEGER NOT NULL, created_at INTEGER NOT NULL,
            UNIQUE(requester_id, recipient_id), CHECK(requester_id != recipient_id))""")
        con.execute("CREATE INDEX IF NOT EXISTS idx_friend_requests_recipient ON friend_requests(recipient_id, created_at)")
        con.execute("""CREATE TABLE IF NOT EXISTS legacy_claims (
            token_hash TEXT PRIMARY KEY, account_id INTEGER NOT NULL,
            FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE)""")
        con.execute("UPDATE messages SET body_enc=1 WHERE body LIKE 'enc1:%' AND CAST(body_enc AS INTEGER)=0")
        con.execute("CREATE UNIQUE INDEX IF NOT EXISTS idx_participant_nick ON participants(room, nickname_key)")
        con.execute("CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room, id)")
        con.execute("DELETE FROM account_sessions WHERE expires_at<?", (now_ts(),))


def migrate_bodies():
    with LOCK, db_connection() as con:
        rows = con.execute("SELECT id,body FROM messages WHERE CAST(body_enc AS INTEGER)=0").fetchall()
        for mid, body in rows:
            con.execute("UPDATE messages SET body=?,body_enc=1 WHERE id=?", (encrypt_body(body), mid))
        return len(rows)


def json_bytes(value):
    return json.dumps(value, ensure_ascii=False).encode("utf-8")


def account_payload(row):
    return {"nickname": row[1], "avatar_id": row[2], "color_id": row[3], "registered_nickname": row[-1] or row[1]}


@contextmanager
def db_connection():
    con = sqlite3.connect(DB, timeout=5)
    try:
        yield con
        con.commit()
    except BaseException:
        con.rollback()
        raise
    finally:
        con.close()


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *_args):
        return

    # Trust Nginx's overwritten X-Real-IP only when the TCP peer is loopback.
    def client_key(self):
        try:
            peer = ipaddress.ip_address(self.client_address[0])
        except ValueError:
            return self.client_address[0]
        if peer.is_loopback:
            try:
                forwarded = ipaddress.ip_address(self.headers.get("X-Real-IP", "").strip())
                return str(getattr(forwarded, "ipv4_mapped", None) or forwarded)
            except ValueError:
                pass
        return str(getattr(peer, "ipv4_mapped", None) or peer)

    def ip_digest(self):
        # Store a keyed pseudonym, never the raw address.
        return hmac.new(KEY, b"messenger-registration-ip:v1:" + self.client_key().encode(), hashlib.sha256).hexdigest()

    def allow(self, bucket, limit, window):
        key = (self.client_key(), bucket)
        current = time.monotonic()
        with LOCK:
            attempts = [stamp for stamp in RATE.get(key, []) if current - stamp < window]
            if len(attempts) >= limit:
                RATE[key] = attempts
                return False
            attempts.append(current)
            RATE[key] = attempts
            return True

    def send_json(self, value, status=200, headers=()):
        data = json_bytes(value)
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Cache-Control", "no-store")
        self.send_header("X-Content-Type-Options", "nosniff")
        for name, val in headers:
            self.send_header(name, val)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def read_json(self):
        try:
            size = int(self.headers.get("Content-Length", "0"))
            if size < 0 or size > 16_000:
                return None
            return json.loads(self.rfile.read(size))
        except (ValueError, json.JSONDecodeError):
            return None

    def origin_ok(self):
        return self.headers.get("Origin", "") in ALLOWED_ORIGINS

    def cookie_token(self):
        try:
            cookie = SimpleCookie(self.headers.get("Cookie", ""))
            return cookie["messenger_session"].value if "messenger_session" in cookie else ""
        except (CookieError, KeyError, ValueError):
            return ""

    def current_account(self, con):
        raw = self.cookie_token()
        if not raw:
            return None
        row = con.execute("""SELECT a.id,a.nickname,a.avatar_id,a.color_id,a.token_hash,a.registered_nickname
            FROM account_sessions s JOIN accounts a ON a.id=s.account_id
            WHERE s.token_hash=? AND s.expires_at>?""", (token_hash(raw), now_ts())).fetchone()
        return row

    def session_cookie(self, raw, max_age=SESSION_SECONDS):
        return ("Set-Cookie", f"messenger_session={raw}; Domain=.smlt.lol; Path=/; Max-Age={max_age}; Secure; HttpOnly; SameSite=Lax")

    def has_session(self):
        if not self.cookie_token():
            return False
        with db_connection() as con:
            return self.current_account(con) is not None

    def require_mutation_origin(self):
        if (self.path.startswith("/api/auth/") or self.has_session()) and not self.origin_ok():
            self.send_json({"error": "Запрос с этого сайта запрещён."}, 403)
            return False
        return True

    def identity_hash(self, con, guest_token):
        account = self.current_account(con)
        return account[4] if account else token_hash(guest_token)

    def cleanup(self, con):
        con.execute("DELETE FROM participants WHERE last_seen<?", (now_ts() - STALE_SECONDS,))
        con.execute("DELETE FROM account_sessions WHERE expires_at<?", (now_ts(),))

    def room_row(self, con, code):
        return con.execute("SELECT code,owner_token,name FROM rooms WHERE code=?", (code,)).fetchone()

    def member(self, con, room, identity):
        return con.execute("SELECT nickname FROM participants WHERE room=? AND token_hash=? AND last_seen>=?", (room, identity, now_ts() - STALE_SECONDS)).fetchone()

    def do_GET(self):
        parsed = urlparse(self.path)
        if parsed.path == "/health":
            return self.send_json({"ok": True, "encryption": "aes-256-gcm", "accounts": True})
        if parsed.path == "/api/auth/me":
            with db_connection() as con:
                account = self.current_account(con)
            if not account:
                return self.send_json({"authenticated": False}, 401)
            return self.send_json({"authenticated": True, "account": account_payload(account)})
        if parsed.path == "/api/friends":
            with db_connection() as con:
                account = self.current_account(con)
                if not account:
                    return self.send_json({"error": "Войдите в аккаунт."}, 401)
                friends = con.execute("""SELECT a.nickname,a.registered_nickname,a.avatar_id,a.color_id
                    FROM friendships f JOIN accounts a ON a.id=CASE WHEN f.account_a=? THEN f.account_b ELSE f.account_a END
                    WHERE f.account_a=? OR f.account_b=? ORDER BY a.nickname COLLATE NOCASE""",
                    (account[0], account[0], account[0])).fetchall()
                incoming = con.execute("""SELECT r.id,a.nickname,a.registered_nickname,a.avatar_id,a.color_id
                    FROM friend_requests r JOIN accounts a ON a.id=r.requester_id
                    WHERE r.recipient_id=? ORDER BY r.created_at""", (account[0],)).fetchall()
                outgoing = con.execute("""SELECT a.nickname,a.registered_nickname
                    FROM friend_requests r JOIN accounts a ON a.id=r.recipient_id
                    WHERE r.requester_id=? ORDER BY r.created_at""", (account[0],)).fetchall()
            pack = lambda row: {"nickname": row[0], "registered_nickname": row[1] or row[0], "avatar_id": row[2], "color_id": row[3]}
            return self.send_json({"friends": [pack(row) for row in friends],
                "incoming": [{"id": row[0], **pack(row[1:])} for row in incoming],
                "outgoing": [{"nickname": row[0], "registered_nickname": row[1] or row[0]} for row in outgoing]})
        if parsed.path == "/api/notifications":
            query = parse_qs(parsed.query)
            try:
                after = max(0, int(query.get("after", ["0"])[0]))
            except ValueError:
                return self.send_json({"error": "Invalid notification cursor."}, 400)
            if after > 9223372036854775807:
                return self.send_json({"error": "Invalid notification cursor."}, 400)
            with LOCK, db_connection() as con:
                account = self.current_account(con)
                if not account:
                    return self.send_json({"error": "Войдите в аккаунт."}, 401)
                latest = con.execute("""SELECT COALESCE(MAX(m.id),0) FROM messages m
                    JOIN account_rooms ar ON ar.room=m.room WHERE ar.account_id=?""", (account[0],)).fetchone()[0]
                if query.get("baseline", [""])[0] == "1":
                    return self.send_json({"notifications": [], "cursor": latest})
                rows = con.execute("""SELECT m.id,m.room,r.name,m.nickname,m.body,m.created_at
                    FROM messages m JOIN account_rooms ar ON ar.room=m.room
                    JOIN rooms r ON r.code=m.room
                    WHERE ar.account_id=? AND m.id>? AND m.author_token!=?
                    ORDER BY m.id LIMIT 40""", (account[0], after, account[4])).fetchall()
            notices = [{"id": row[0], "room": row[1], "room_name": row[2], "nickname": row[3],
                "body": decrypt_body(row[4])[:180], "created_at": row[5]} for row in rows]
            cursor = rows[-1][0] if len(rows) >= 40 else max(latest, rows[-1][0] if rows else after)
            return self.send_json({"notifications": notices, "cursor": cursor})
        if parsed.path == "/api/rooms":
            with db_connection() as con:
                account = self.current_account(con)
                if not account:
                    return self.send_json({"error": "Войдите в аккаунт."}, 401)
                rows = con.execute("""SELECT r.code,r.name,r.owner_token FROM account_rooms ar
                    JOIN rooms r ON r.code=ar.room WHERE ar.account_id=? ORDER BY ar.added_at DESC""", (account[0],)).fetchall()
            return self.send_json({"rooms": [{"code": row[0], "name": row[1], "owner": row[2] == account[4]} for row in rows]})
        if parsed.path == "/api/messages":
            query = parse_qs(parsed.query)
            room = clean_text(query.get("room", [""])[0], 128)
            guest_token = self.headers.get("X-Author-Token", "")[:256]
            try:
                after = max(0, int(query.get("after", [0])[0]))
            except ValueError:
                after = 0
            if not room or len(guest_token) < 32:
                return self.send_json({"error": "room and author token are required"}, 400)
            with LOCK, db_connection() as con:
                self.cleanup(con)
                identity = self.identity_hash(con, guest_token)
                rr = self.room_row(con, room)
                if not rr:
                    return self.send_json({"error": "Комната была удалена или не найдена.", "code": "ROOM_NOT_FOUND"}, 404)
                if not self.member(con, room, identity):
                    return self.send_json({"error": "Сначала войдите в комнату.", "code": "NOT_JOINED"}, 403)
                con.execute("UPDATE participants SET last_seen=? WHERE room=? AND token_hash=?", (now_ts(), room, identity))
                rows = con.execute("""SELECT m.id,m.nickname,m.body,m.created_at,m.author_token,m.avatar_id,m.color_id,
                    a.registered_nickname FROM messages m
                    LEFT JOIN accounts a ON a.token_hash=m.author_token
                    WHERE m.room=? AND m.id>? ORDER BY m.id LIMIT 100""", (room, after)).fetchall()
            messages = [{"id": mid, "nickname": nick, "body": decrypt_body(body), "created_at": created, "mine": bool(author and author == identity), "avatar_id": avatar if avatar in AVATARS else DEFAULT_AVATAR, "color_id": color if color in COLORS else DEFAULT_COLOR, "registered_nickname": registered_nick} for mid, nick, body, created, author, avatar, color, registered_nick in rows]
            return self.send_json({"messages": messages, "room": {"code": rr[0], "name": rr[2], "owner": rr[1] == identity}})
        self.send_error(404)

    def do_POST(self):
        if not self.require_mutation_origin():
            return
        if self.path == "/api/friends/requests":
            if not self.allow("friend_request", 12, 300):
                return self.send_json({"error": "Слишком много запросов в друзья. Попробуйте позже."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw_nickname = str(payload.get("nickname", ""))
            nickname = nick_value(raw_nickname)
            if len(raw_nickname.strip()) > MAX_NICK:
                return self.send_json({"error": "Ник не должен быть длиннее 32 символов."}, 400)
            if not nickname or not nick_key(nickname):
                return self.send_json({"error": "Укажите ник пользователя."}, 400)
            with LOCK, db_connection() as con:
                con.execute("BEGIN IMMEDIATE")
                account = self.current_account(con)
                if not account:
                    con.rollback()
                    return self.send_json({"error": "Войдите в аккаунт."}, 401)
                target = con.execute("SELECT id,nickname FROM accounts WHERE nickname_key=?", (nick_key(nickname),)).fetchone()
                if not target:
                    con.rollback()
                    return self.send_json({"error": "Аккаунт с таким текущим ником не найден."}, 404)
                if target[0] == account[0]:
                    con.rollback()
                    return self.send_json({"error": "Нельзя добавить самого себя."}, 400)
                a, b = sorted((account[0], target[0]))
                if con.execute("SELECT 1 FROM friendships WHERE account_a=? AND account_b=?", (a, b)).fetchone():
                    con.rollback()
                    return self.send_json({"error": "Этот пользователь уже в друзьях."}, 409)
                reverse = con.execute("SELECT id FROM friend_requests WHERE requester_id=? AND recipient_id=?", (target[0], account[0])).fetchone()
                if reverse:
                    con.execute("INSERT OR IGNORE INTO friendships(account_a,account_b,created_at) VALUES(?,?,?)", (a, b, now_ts()))
                    con.execute("DELETE FROM friend_requests WHERE (requester_id=? AND recipient_id=?) OR (requester_id=? AND recipient_id=?)", (account[0], target[0], target[0], account[0]))
                    con.commit()
                    return self.send_json({"ok": True, "status": "friends"}, 201)
                try:
                    con.execute("INSERT INTO friend_requests(requester_id,recipient_id,created_at) VALUES(?,?,?)", (account[0], target[0], now_ts()))
                except sqlite3.IntegrityError:
                    con.rollback()
                    return self.send_json({"error": "Запрос этому пользователю уже отправлен."}, 409)
                con.commit()
            return self.send_json({"ok": True, "status": "pending"}, 201)
        if self.path.startswith("/api/friends/requests/"):
            suffix = self.path[len("/api/friends/requests/"):]
            try:
                request_id = int(suffix)
            except ValueError:
                return self.send_json({"error": "Некорректный запрос."}, 400)
            if request_id < 1 or request_id > 9223372036854775807:
                return self.send_json({"error": "Некорректный запрос."}, 400)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            action = payload.get("action")
            if action not in {"accept", "decline"}:
                return self.send_json({"error": "Выберите действие с запросом."}, 400)
            with LOCK, db_connection() as con:
                con.execute("BEGIN IMMEDIATE")
                account = self.current_account(con)
                request = con.execute("SELECT requester_id,recipient_id FROM friend_requests WHERE id=?", (request_id,)).fetchone() if account else None
                if not request or request[1] != account[0]:
                    con.rollback()
                    return self.send_json({"error": "Запрос не найден."}, 404)
                if action == "accept":
                    a, b = sorted((request[0], request[1]))
                    con.execute("INSERT OR IGNORE INTO friendships(account_a,account_b,created_at) VALUES(?,?,?)", (a, b, now_ts()))
                con.execute("DELETE FROM friend_requests WHERE id=?", (request_id,))
                con.commit()
            return self.send_json({"ok": True, "status": "friends" if action == "accept" else "declined"})
        if self.path == "/api/friends/remove":
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw_nickname = str(payload.get("nickname", ""))
            nickname = nick_value(raw_nickname)
            if len(raw_nickname.strip()) > MAX_NICK:
                return self.send_json({"error": "Ник не должен быть длиннее 32 символов."}, 400)
            with LOCK, db_connection() as con:
                account = self.current_account(con)
                if not account:
                    return self.send_json({"error": "Войдите в аккаунт."}, 401)
                target = con.execute("SELECT id FROM accounts WHERE nickname_key=?", (nick_key(nickname),)).fetchone()
                if not target or target[0] == account[0]:
                    return self.send_json({"error": "Друг не найден."}, 404)
                a, b = sorted((account[0], target[0]))
                con.execute("DELETE FROM friendships WHERE account_a=? AND account_b=?", (a, b))
            return self.send_json({"ok": True})
        if self.path == "/api/auth/register":
            if not self.allow("register", 5, 300):
                return self.send_json({"error": "Слишком много попыток регистрации. Попробуйте позже."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw_nick = str(payload.get("nickname", ""))
            nickname = nick_value(raw_nick)
            password = payload.get("password", "")
            if len(raw_nick.strip()) > MAX_NICK or len(nickname) < 3 or not nick_key(nickname) or not valid_password(password):
                return self.send_json({"error": "Укажи ник длиной 3–32 символа и пароль длиной 12–128 символов."}, 400)
            salt = secrets.token_bytes(16)
            recovery = new_recovery_code()
            recovery_salt = secrets.token_bytes(16)
            canonical_hash = token_hash(secrets.token_urlsafe(32))
            session = secrets.token_urlsafe(32)
            ip_hash = self.ip_digest()
            with LOCK, db_connection() as con:
                con.execute("BEGIN IMMEDIATE")
                if con.execute("SELECT 1 FROM accounts WHERE nickname_key=?", (nick_key(nickname),)).fetchone():
                    con.rollback()
                    return self.send_json({"error": "Этот ник уже занят."}, 409)
                count = con.execute("SELECT count(*) FROM registration_ips WHERE ip_hash=?", (ip_hash,)).fetchone()[0]
                if count >= 3:
                    con.rollback()
                    return self.send_json({"error": "С этого IP уже создано 3 аккаунта."}, 429)
                cur = con.execute("""INSERT INTO accounts(nickname,registered_nickname,nickname_key,password_salt,password_hash,
                    recovery_salt,recovery_hash,token_hash,avatar_id,color_id,created_at)
                    VALUES(?,?,?,?,?,?,?,?,?,?,?)""", (nickname, nickname, nick_key(nickname), salt, password_hash(password, salt), recovery_salt, password_hash(recovery_value(recovery), recovery_salt), canonical_hash, DEFAULT_AVATAR, DEFAULT_COLOR, now_ts()))
                account_id = cur.lastrowid
                con.execute("INSERT INTO registration_ips(ip_hash,account_id,created_at) VALUES(?,?,?)", (ip_hash, account_id, now_ts()))
                con.execute("INSERT INTO account_sessions(token_hash,account_id,expires_at,created_at) VALUES(?,?,?,?)", (token_hash(session), account_id, now_ts() + SESSION_SECONDS, now_ts()))
                con.commit()
                account = con.execute("SELECT id,nickname,avatar_id,color_id,token_hash,registered_nickname FROM accounts WHERE id=?", (account_id,)).fetchone()
            return self.send_json({"ok": True, "account": account_payload(account), "recovery_code": recovery}, 201, (self.session_cookie(session),))

        if self.path == "/api/auth/login":
            if not self.allow("login", 12, 300):
                return self.send_json({"error": "Слишком много попыток входа. Попробуйте позже."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            nickname = nick_value(payload.get("nickname", ""))
            password = payload.get("password", "")
            with LOCK, db_connection() as con:
                account = con.execute("SELECT id,nickname,avatar_id,color_id,token_hash,password_salt,password_hash,registered_nickname FROM accounts WHERE nickname_key=?", (nick_key(nickname),)).fetchone()
                valid_input = isinstance(password, str) and len(password) <= 128
                safe_password = password if valid_input else "invalid"
                salt = account[5] if account else b"messenger-fake-salt"
                candidate = password_hash(safe_password, salt)
                valid = bool(account and valid_input and hmac.compare_digest(candidate, account[6]))
                if not valid:
                    return self.send_json({"error": "Неверный ник или пароль."}, 401)
                session = secrets.token_urlsafe(32)
                con.execute("DELETE FROM account_sessions WHERE expires_at<?", (now_ts(),))
                con.execute("INSERT INTO account_sessions(token_hash,account_id,expires_at,created_at) VALUES(?,?,?,?)", (token_hash(session), account[0], now_ts() + SESSION_SECONDS, now_ts()))
            return self.send_json({"ok": True, "account": account_payload(account)}, headers=(self.session_cookie(session),))

        if self.path == "/api/auth/recover":
            if not self.allow("recover", 5, 900):
                return self.send_json({"error": "Слишком много попыток восстановления. Попробуйте позже."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            nickname = nick_value(payload.get("nickname", ""))
            recovery = recovery_value(payload.get("recovery_code", ""))
            password = payload.get("password", "")
            if not valid_password(password):
                return self.send_json({"error": "Новый пароль должен содержать от 12 до 128 символов."}, 400)
            with LOCK, db_connection() as con:
                con.execute("BEGIN IMMEDIATE")
                account = con.execute("SELECT id,recovery_salt,recovery_hash FROM accounts WHERE nickname_key=?", (nick_key(nickname),)).fetchone()
                salt = account[1] if account else b"messenger-fake-salt"
                candidate = password_hash(recovery if len(recovery) == 32 else "invalid", salt)
                if not account or len(recovery) != 32 or not hmac.compare_digest(candidate, account[2]):
                    con.rollback()
                    return self.send_json({"error": "Ник или код восстановления не подошёл."}, 401)
                new_salt = secrets.token_bytes(16)
                new_recovery = new_recovery_code()
                new_recovery_salt = secrets.token_bytes(16)
                con.execute("UPDATE accounts SET password_salt=?,password_hash=?,recovery_salt=?,recovery_hash=? WHERE id=?", (new_salt, password_hash(password, new_salt), new_recovery_salt, password_hash(new_recovery, new_recovery_salt), account[0]))
                con.execute("DELETE FROM account_sessions WHERE account_id=?", (account[0],))
                session = secrets.token_urlsafe(32)
                con.execute("INSERT INTO account_sessions(token_hash,account_id,expires_at,created_at) VALUES(?,?,?,?)", (token_hash(session), account[0], now_ts() + SESSION_SECONDS, now_ts()))
                con.commit()
                profile = con.execute("SELECT nickname,avatar_id,color_id,registered_nickname FROM accounts WHERE id=?", (account[0],)).fetchone()
            return self.send_json({"ok": True, "account": {"nickname": profile[0], "avatar_id": profile[1], "color_id": profile[2], "registered_nickname": profile[3]}, "recovery_code": new_recovery}, headers=(self.session_cookie(session),))

        if self.path == "/api/auth/logout":
            raw = self.cookie_token()
            if raw:
                with db_connection() as con:
                    con.execute("DELETE FROM account_sessions WHERE token_hash=?", (token_hash(raw),))
            return self.send_json({"ok": True}, headers=(self.session_cookie("", 0),))

        if self.path == "/api/auth/profile":
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw_nick = str(payload.get("nickname", ""))
            nickname = nick_value(raw_nick)
            avatar = str(payload.get("avatar_id", ""))
            color = str(payload.get("color_id", ""))
            if len(raw_nick.strip()) > MAX_NICK or len(nickname) < 3 or not nick_key(nickname) or not valid_profile(avatar, color):
                return self.send_json({"error": "Проверь ник, аватар и цвет."}, 400)
            with LOCK, db_connection() as con:
                account = self.current_account(con)
                if not account:
                    return self.send_json({"error": "Сначала войди в аккаунт."}, 401)
                try:
                    con.execute("UPDATE accounts SET nickname=?,nickname_key=?,avatar_id=?,color_id=? WHERE id=?", (nickname, nick_key(nickname), avatar, color, account[0]))
                except sqlite3.IntegrityError:
                    return self.send_json({"error": "Этот ник уже занят."}, 409)
            return self.send_json({"ok": True, "account": {"nickname": nickname, "avatar_id": avatar, "color_id": color, "registered_nickname": account[5]}})

        if self.path == "/api/auth/claim":
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            guest_token = str(payload.get("author_token", ""))[:256]
            if len(guest_token) < 32:
                return self.send_json({"error": "Не найден профиль этого устройства."}, 400)
            with LOCK, db_connection() as con:
                con.execute("BEGIN IMMEDIATE")
                account = self.current_account(con)
                if not account:
                    con.rollback()
                    return self.send_json({"error": "Сначала войди в аккаунт."}, 401)
                old_hash, canonical = token_hash(guest_token), account[4]
                owner = con.execute("SELECT account_id FROM legacy_claims WHERE token_hash=?", (old_hash,)).fetchone()
                if owner and owner[0] != account[0]:
                    con.rollback()
                    return self.send_json({"error": "Этот старый профиль уже привязан к другому аккаунту."}, 409)
                if old_hash != canonical:
                    con.execute("INSERT OR IGNORE INTO legacy_claims(token_hash,account_id) VALUES(?,?)", (old_hash, account[0]))
                    con.execute("INSERT OR IGNORE INTO account_rooms(account_id,room,added_at) SELECT ?,code,? FROM rooms WHERE owner_token=?", (account[0], now_ts(), old_hash))
                    con.execute("INSERT OR IGNORE INTO account_rooms(account_id,room,added_at) SELECT ?,room,? FROM participants WHERE token_hash=?", (account[0], now_ts(), old_hash))
                    con.execute("UPDATE rooms SET owner_token=? WHERE owner_token=?", (canonical, old_hash))
                    con.execute("UPDATE messages SET author_token=? WHERE author_token=?", (canonical, old_hash))
                    con.execute("DELETE FROM participants WHERE token_hash=? AND EXISTS (SELECT 1 FROM participants p2 WHERE p2.room=participants.room AND p2.token_hash=?)", (old_hash, canonical))
                    con.execute("UPDATE participants SET token_hash=? WHERE token_hash=?", (canonical, old_hash))
                con.commit()
            return self.send_json({"ok": True})

        if self.path == "/api/rooms":
            if not self.allow("create", 5, 60):
                return self.send_json({"error": "Слишком много новых комнат. Попробуйте позже."}, 429)
            payload = self.read_json()
            guest_token = str((payload or {}).get("author_token", ""))[:256] if isinstance(payload, dict) else ""
            if len(guest_token) < 32:
                return self.send_json({"error": "author token is required"}, 400)
            code = secrets.token_urlsafe(32)
            with LOCK, db_connection() as con:
                identity = self.identity_hash(con, guest_token)
                con.execute("INSERT INTO rooms(code,owner_token,name) VALUES(?,?,?)", (code, identity, "Новая комната"))
                account = self.current_account(con)
                if account:
                    con.execute("INSERT INTO account_rooms(account_id,room,added_at) VALUES(?,?,?)", (account[0], code, now_ts()))
            return self.send_json({"ok": True, "code": code, "name": "Новая комната"})

        if self.path == "/api/rooms/join":
            if not self.allow("join", 30, 60):
                return self.send_json({"error": "Слишком много попыток входа. Попробуйте позже."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw = str(payload.get("nickname", ""))
            room = clean_text(payload.get("room"), 128)
            nick = nick_value(raw)
            guest_token = str(payload.get("author_token", ""))[:256]
            avatar = str(payload.get("avatar_id", DEFAULT_AVATAR))
            color = str(payload.get("color_id", DEFAULT_COLOR))
            if len(raw.strip()) > MAX_NICK or not room or not nick or not nick_key(nick) or len(guest_token) < 32 or not valid_profile(avatar, color):
                return self.send_json({"error": "Введите корректный профиль и код комнаты."}, 400)
            with LOCK, db_connection() as con:
                self.cleanup(con)
                account = self.current_account(con)
                identity = account[4] if account else token_hash(guest_token)
                if account:
                    nick, avatar, color = account[1], account[2], account[3]
                rr = self.room_row(con, room)
                if not rr:
                    legacy = con.execute("SELECT 1 FROM messages WHERE room=? LIMIT 1", (room,)).fetchone()
                    if not legacy:
                        return self.send_json({"error": "Комната не найдена.", "code": "ROOM_NOT_FOUND"}, 404)
                    con.execute("INSERT OR IGNORE INTO rooms(code,owner_token,name) VALUES(?,?,?)", (room, "", "Новая комната"))
                    rr = self.room_row(con, room)
                existing = con.execute("SELECT token_hash FROM participants WHERE room=? AND nickname_key=?", (room, nick_key(nick))).fetchone()
                if existing and existing[0] != identity:
                    return self.send_json({"error": "Этот ник уже используется в комнате.", "code": "NICK_BUSY"}, 409)
                con.execute("""INSERT INTO participants(room,token_hash,nickname,nickname_key,last_seen,avatar_id,color_id)
                    VALUES(?,?,?,?,?,?,?) ON CONFLICT(room,token_hash) DO UPDATE SET
                    nickname=excluded.nickname,nickname_key=excluded.nickname_key,last_seen=excluded.last_seen,
                    avatar_id=excluded.avatar_id,color_id=excluded.color_id""", (room, identity, nick, nick_key(nick), now_ts(), avatar, color))
                if account:
                    con.execute("INSERT OR IGNORE INTO account_rooms(account_id,room,added_at) VALUES(?,?,?)", (account[0], room, now_ts()))
            return self.send_json({"ok": True, "room": {"code": rr[0], "name": rr[2], "owner": rr[1] == identity}})

        if self.path == "/api/messages":
            if not self.allow("message", 12, 10):
                return self.send_json({"error": "Слишком много сообщений. Подождите немного."}, 429)
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            raw = str(payload.get("nickname", ""))
            raw_body = str(payload.get("body", ""))
            room = clean_text(payload.get("room"), 128)
            nick = nick_value(raw)
            body = clean_text(raw_body, MAX_MESSAGE)
            guest_token = str(payload.get("author_token", ""))[:256]
            avatar = str(payload.get("avatar_id", DEFAULT_AVATAR))
            color = str(payload.get("color_id", DEFAULT_COLOR))
            if len(raw.strip()) > MAX_NICK or len(raw_body.strip()) > MAX_MESSAGE or not room or not nick or not body or len(guest_token) < 32 or not valid_profile(avatar, color):
                return self.send_json({"error": "Профиль и сообщение заполнены неверно."}, 400)
            with LOCK, db_connection() as con:
                self.cleanup(con)
                account = self.current_account(con)
                identity = account[4] if account else token_hash(guest_token)
                if account:
                    nick, avatar, color = account[1], account[2], account[3]
                rr = self.room_row(con, room)
                if not rr:
                    return self.send_json({"error": "Комната была удалена или не найдена.", "code": "ROOM_NOT_FOUND"}, 404)
                if not self.member(con, room, identity):
                    return self.send_json({"error": "Сначала войдите в комнату.", "code": "NOT_JOINED"}, 403)
                cur = con.execute("INSERT INTO messages(room,nickname,body,author_token,avatar_id,color_id,body_enc) VALUES(?,?,?,?,?,?,1)", (room, nick, encrypt_body(body), identity, avatar, color))
                message_id = cur.lastrowid
            return self.send_json({"ok": True, "id": message_id})

        if self.path == "/api/rooms/leave":
            payload = self.read_json()
            payload = payload if isinstance(payload, dict) else {}
            room = clean_text(payload.get("room"), 128)
            guest_token = str(payload.get("author_token", ""))[:256]
            with LOCK, db_connection() as con:
                identity = self.identity_hash(con, guest_token)
                con.execute("DELETE FROM participants WHERE room=? AND token_hash=?", (room, identity))
                account = self.current_account(con)
                if account:
                    con.execute("DELETE FROM account_rooms WHERE account_id=? AND room=?", (account[0], room))
            return self.send_json({"ok": True})

        self.send_error(404)

    def do_PATCH(self):
        if not self.require_mutation_origin():
            return
        if not self.path.startswith("/api/rooms/"):
            return self.send_error(404)
        room = clean_text(self.path[len("/api/rooms/"):], 128)
        payload = self.read_json()
        payload = payload if isinstance(payload, dict) else {}
        guest_token = str(payload.get("author_token", ""))[:256]
        name = clean_text(payload.get("name"), MAX_ROOM_NAME)
        if not name:
            return self.send_json({"error": "Название не может быть пустым."}, 400)
        with LOCK, db_connection() as con:
            identity = self.identity_hash(con, guest_token)
            rr = self.room_row(con, room)
            if not rr:
                return self.send_json({"error": "Комната не найдена.", "code": "ROOM_NOT_FOUND"}, 404)
            if rr[1] != identity:
                return self.send_json({"error": "Переименовать может только создатель."}, 403)
            con.execute("UPDATE rooms SET name=? WHERE code=?", (name, room))
        return self.send_json({"ok": True, "name": name})

    def do_DELETE(self):
        if not self.require_mutation_origin():
            return
        if self.path.startswith("/api/rooms/"):
            room = clean_text(self.path[len("/api/rooms/"):], 128)
            payload = self.read_json()
            guest_token = str((payload or {}).get("author_token", ""))[:256] if isinstance(payload, dict) else ""
            with LOCK, db_connection() as con:
                identity = self.identity_hash(con, guest_token)
                rr = self.room_row(con, room)
                if not rr:
                    return self.send_json({"error": "Комната уже удалена.", "code": "ROOM_NOT_FOUND"}, 404)
                if rr[1] != identity:
                    return self.send_json({"error": "Удалить может только создатель."}, 403)
                con.execute("DELETE FROM messages WHERE room=?", (room,))
                con.execute("DELETE FROM participants WHERE room=?", (room,))
                con.execute("DELETE FROM account_rooms WHERE room=?", (room,))
                con.execute("DELETE FROM rooms WHERE code=?", (room,))
            return self.send_json({"ok": True})
        prefix = "/api/messages/"
        if not self.path.startswith(prefix):
            return self.send_error(404)
        try:
            mid = int(self.path[len(prefix):])
        except ValueError:
            return self.send_json({"error": "invalid message id"}, 400)
        payload = self.read_json()
        payload = payload if isinstance(payload, dict) else {}
        room = clean_text(payload.get("room"), 128)
        guest_token = str(payload.get("author_token", ""))[:256]
        with LOCK, db_connection() as con:
            identity = self.identity_hash(con, guest_token)
            cur = con.execute("DELETE FROM messages WHERE id=? AND room=? AND author_token=?", (mid, room, identity))
        if not cur.rowcount:
            return self.send_json({"error": "Сообщение не найдено или оно не ваше."}, 403)
        return self.send_json({"ok": True})


if __name__ == "__main__":
    KEY = key_bytes()
    db_init()
    migrated = migrate_bodies()
    print(f"Encrypted message bodies migrated: {migrated}", flush=True)
    port = int(os.environ.get("MESSENGER_PORT", "8787"))
    print(f"Messenger listening on 127.0.0.1:{port}", flush=True)
    ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()
