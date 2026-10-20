import importlib.util
import sqlite3
import tempfile
import json
import secrets
import threading
import unittest
import urllib.error
import urllib.request
from pathlib import Path


SERVER_PATH = Path(__file__).with_name("server.py")
SPEC = importlib.util.spec_from_file_location("messenger_server_test_target", SERVER_PATH)
server = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(server)


class AccountRegistrationNicknameMigrationTests(unittest.TestCase):
    def setUp(self):
        self.tempdir = tempfile.TemporaryDirectory(prefix="messenger-account-test-")
        self.original_db = server.DB
        self.original_data_dir = server.DATA_DIR
        server.DATA_DIR = Path(self.tempdir.name)
        server.DB = server.DATA_DIR / "messages.sqlite3"
        with server.db_connection() as con:
            con.execute("""CREATE TABLE accounts (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                nickname TEXT NOT NULL, nickname_key TEXT NOT NULL UNIQUE,
                password_salt BLOB NOT NULL, password_hash BLOB NOT NULL,
                recovery_salt BLOB NOT NULL, recovery_hash BLOB NOT NULL,
                token_hash TEXT NOT NULL UNIQUE,
                avatar_id TEXT NOT NULL DEFAULT 'avatar1',
                color_id TEXT NOT NULL DEFAULT 'violet', created_at INTEGER NOT NULL)""")
            con.execute("""INSERT INTO accounts
                (nickname,nickname_key,password_salt,password_hash,recovery_salt,
                 recovery_hash,token_hash,created_at)
                VALUES('FirstNick','firstnick',x'00',x'00',x'00',x'00','account-token',1)""")
        server.db_init()

    def tearDown(self):
        server.DB = self.original_db
        server.DATA_DIR = self.original_data_dir
        self.tempdir.cleanup()

    def test_migration_preserves_legacy_current_nickname_and_is_idempotent(self):
        server.db_init()
        with server.db_connection() as con:
            columns = {row[1] for row in con.execute("PRAGMA table_info(accounts)")}
            self.assertIn("registered_nickname", columns)
            self.assertEqual(
                con.execute("SELECT registered_nickname FROM accounts WHERE id=1").fetchone()[0],
                "FirstNick",
            )

    def test_profile_update_does_not_change_public_registration_nickname(self):
        with server.db_connection() as con:
            con.execute("UPDATE accounts SET nickname='CurrentNick' WHERE id=1")
            con.execute("INSERT INTO rooms(code,owner_token,name) VALUES('room','owner','Room')")
            con.execute("""INSERT INTO messages
                (room,nickname,body,author_token) VALUES('room','CurrentNick','body','account-token')""")
            registered = con.execute("""SELECT a.registered_nickname FROM messages m
                LEFT JOIN accounts a ON a.token_hash=m.author_token WHERE m.room='room'""").fetchone()[0]
        self.assertEqual(registered, "FirstNick")


class FriendsAndNotificationsApiTests(unittest.TestCase):
    def setUp(self):
        self.tempdir = tempfile.TemporaryDirectory(prefix="messenger-social-test-")
        self.original_db = server.DB
        self.original_data_dir = server.DATA_DIR
        self.original_key = server.KEY
        server.DATA_DIR = Path(self.tempdir.name)
        server.DB = server.DATA_DIR / "messages.sqlite3"
        server.KEY = secrets.token_bytes(32)
        server.db_init()
        self.sessions = {}
        with server.db_connection() as con:
            for account_id, nickname in ((1, "FirstUser"), (2, "SecondUser")):
                identity = server.token_hash(f"identity-{account_id}")
                session = f"session-{account_id}-" + secrets.token_urlsafe(32)
                self.sessions[account_id] = session
                con.execute("""INSERT INTO accounts
                    (id,nickname,registered_nickname,nickname_key,password_salt,password_hash,
                     recovery_salt,recovery_hash,token_hash,avatar_id,color_id,created_at)
                    VALUES(?,?,?,?,?,?,?,?,?,?,?,?)""",
                    (account_id,nickname,nickname,server.nick_key(nickname),b"s",b"h",b"r",b"h",identity,"avatar1","violet",1))
                con.execute("INSERT INTO account_sessions(token_hash,account_id,expires_at,created_at) VALUES(?,?,?,?)",
                    (server.token_hash(session),account_id,server.now_ts()+3600,server.now_ts()))
            con.execute("INSERT INTO rooms(code,owner_token,name) VALUES(?,?,?)",("shared","owner","Shared room"))
            con.execute("INSERT INTO rooms(code,owner_token,name) VALUES(?,?,?)",("private-to-second","owner","Other room"))
            con.execute("INSERT INTO account_rooms(account_id,room,added_at) VALUES(1,'shared',1)")
            con.execute("INSERT INTO account_rooms(account_id,room,added_at) VALUES(2,'shared',1)")
            con.execute("INSERT INTO account_rooms(account_id,room,added_at) VALUES(2,'private-to-second',1)")
        self.httpd = server.ThreadingHTTPServer(("127.0.0.1", 0), server.Handler)
        self.thread = threading.Thread(target=self.httpd.serve_forever, daemon=True)
        self.thread.start()
        self.base = f"http://127.0.0.1:{self.httpd.server_port}"

    def tearDown(self):
        self.httpd.shutdown()
        self.httpd.server_close()
        self.thread.join(timeout=2)
        server.DB = self.original_db
        server.DATA_DIR = self.original_data_dir
        server.KEY = self.original_key
        self.tempdir.cleanup()

    def call(self, path, account_id=None, method="GET", payload=None):
        headers = {"Origin": "https://chat.smlt.lol"}
        if account_id:
            headers["Cookie"] = f"messenger_session={self.sessions[account_id]}"
        data = json.dumps(payload).encode() if payload is not None else None
        if data is not None:
            headers["Content-Type"] = "application/json"
        request = urllib.request.Request(self.base + path, data=data, headers=headers, method=method)
        try:
            response = urllib.request.urlopen(request, timeout=5)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            return response.code, json.loads(response.read())

    def test_friend_request_requires_acceptance_and_creates_mutual_friendship(self):
        status, result = self.call("/api/friends/requests", 1, "POST", {"nickname": "SecondUser"})
        self.assertEqual(status, 201)
        self.assertEqual(result["status"], "pending")
        status, duplicate = self.call("/api/friends/requests", 1, "POST", {"nickname": "SecondUser"})
        self.assertEqual(status, 409)
        status, inbox = self.call("/api/friends", 2)
        self.assertEqual(status, 200)
        self.assertEqual(inbox["incoming"][0]["nickname"], "FirstUser")
        status, _ = self.call(f"/api/friends/requests/{inbox['incoming'][0]['id']}", 2, "POST", {"action": "accept"})
        self.assertEqual(status, 200)
        status, friends = self.call("/api/friends", 1)
        self.assertEqual(status, 200)
        self.assertEqual([friend["nickname"] for friend in friends["friends"]], ["SecondUser"])

    def test_friend_api_requires_account_session(self):
        status, _ = self.call("/api/friends")
        self.assertEqual(status, 401)
        status, _ = self.call("/api/notifications?baseline=1")
        self.assertEqual(status, 401)
        status, _ = self.call("/api/friends/requests", method="POST", payload={"nickname": "SecondUser"})
        self.assertEqual(status, 401)

    def test_notifications_only_include_new_messages_in_account_rooms(self):
        status, baseline = self.call("/api/notifications?baseline=1", 1)
        self.assertEqual(status, 200)
        own = server.token_hash("identity-1")
        other = server.token_hash("identity-2")
        with server.db_connection() as con:
            con.execute("INSERT INTO messages(room,nickname,body,author_token) VALUES(?,?,?,?)",
                ("shared", "SecondUser", server.encrypt_body("hello"), other))
            con.execute("INSERT INTO messages(room,nickname,body,author_token) VALUES(?,?,?,?)",
                ("shared", "FirstUser", server.encrypt_body("my own message"), own))
            con.execute("INSERT INTO messages(room,nickname,body,author_token) VALUES(?,?,?,?)",
                ("private-to-second", "SecondUser", server.encrypt_body("not in your room"), other))
        status, result = self.call(f"/api/notifications?after={baseline['cursor']}", 1)
        self.assertEqual(status, 200)
        self.assertEqual(len(result["notifications"]), 1)
        self.assertEqual(result["notifications"][0]["body"], "hello")
        self.assertEqual(result["notifications"][0]["room"], "shared")


if __name__ == "__main__":
    unittest.main()
