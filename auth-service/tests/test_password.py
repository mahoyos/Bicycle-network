def test_password_recovery_existing_email(client, registered_user, mock_email):
    """Recovery request with existing email returns 200 and sends email."""
    response = client.post("/auth/password-recovery", json={
        "email": registered_user["email"]
    })
    assert response.status_code == 200
    assert "message" in response.json()
    mock_email.assert_called_once()


def test_password_recovery_nonexistent_email(client, mock_email):
    """Recovery request with unknown email still returns 200 (no enumeration)."""
    response = client.post("/auth/password-recovery", json={
        "email": "unknown@example.com"
    })
    assert response.status_code == 200
    mock_email.assert_not_called()


def test_password_reset_success(client, registered_user, mock_email):
    """Password can be reset with a valid recovery token."""
    client.post("/auth/password-recovery", json={
        "email": registered_user["email"]
    })

    from tests.conftest import TestingSessionLocal
    from app.models.user import User
    db = TestingSessionLocal()
    user = db.query(User).filter(User.email == registered_user["email"]).first()
    token = user.reset_token
    db.close()

    response = client.post("/auth/password-reset", json={
        "token": token,
        "new_password": "newpassword456"
    })
    assert response.status_code == 200

    response = client.post("/auth/login", json={
        "email": registered_user["email"],
        "password": "newpassword456"
    })
    assert response.status_code == 200


def test_password_reset_invalid_token(client):
    """Password reset fails with an invalid token."""
    response = client.post("/auth/password-reset", json={
        "token": "invalid-token-xyz",
        "new_password": "newpassword456"
    })
    assert response.status_code == 400


def test_password_reset_token_invalidated_after_use(client, registered_user, mock_email):
    """Recovery token cannot be used twice."""
    client.post("/auth/password-recovery", json={
        "email": registered_user["email"]
    })

    from tests.conftest import TestingSessionLocal
    from app.models.user import User
    db = TestingSessionLocal()
    user = db.query(User).filter(User.email == registered_user["email"]).first()
    token = user.reset_token
    db.close()

    client.post("/auth/password-reset", json={
        "token": token,
        "new_password": "newpassword456"
    })

    response = client.post("/auth/password-reset", json={
        "token": token,
        "new_password": "anotherpassword789"
    })
    assert response.status_code == 400