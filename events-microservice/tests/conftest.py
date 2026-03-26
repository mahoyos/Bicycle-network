from app.database import create_tables
from app import main, database, models
import os
import sys

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

# Add project root to PYTHONPATH so `app` package can be imported from tests.
ROOT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
if ROOT_DIR not in sys.path:
    sys.path.insert(0, ROOT_DIR)


@pytest.fixture(scope="session")
def test_db_engine():
    engine = create_engine("sqlite:///:memory:",
                           connect_args={"check_same_thread": False})
    TestingSessionLocal = sessionmaker(
        autocommit=False, autoflush=False, bind=engine)
    database.engine = engine
    database.SessionLocal = TestingSessionLocal

    create_tables()

    yield TestingSessionLocal


@pytest.fixture(scope="module")
def client(test_db_engine):
    return TestClient(main.app)
