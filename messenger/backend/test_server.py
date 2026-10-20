import importlib.util
import sqlite3
import tempfile
import unittest
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
        with sqlite3.connect(server.DB) as con:
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
        with sqlite3.connect(server.DB) as con:
            columns = {row[1] for row in con.execute("PRAGMA table_info(accounts)")}
            self.assertIn("registered_nickname", columns)
            self.assertEqual(
                con.execute("SELECT registered_nickname FROM accounts WHERE id=1").fetchone()[0],
                "FirstNick",
            )

    def test_profile_update_does_not_change_public_registration_nickname(self):
        with sqlite3.connect(server.DB) as con:
            con.execute("UPDATE accounts SET nickname='CurrentNick' WHERE id=1")
            con.execute("INSERT INTO rooms(code,owner_token,name) VALUES('room','owner','Room')")
            con.execute("""INSERT INTO messages
                (room,nickname,body,author_token) VALUES('room','CurrentNick','body','account-token')""")
            registered = con.execute("""SELECT a.registered_nickname FROM messages m
                LEFT JOIN accounts a ON a.token_hash=m.author_token WHERE m.room='room'""").fetchone()[0]
        self.assertEqual(registered, "FirstNick")


if __name__ == "__main__":
    unittest.main()
