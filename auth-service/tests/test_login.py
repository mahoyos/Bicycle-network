def test_login_success(client, registered_user):
    """A registered user can log in with valid credentials."""
    response = client.post("/auth/login", json=registered_user)
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data
    assert "refresh_token" in data
    assert data["token_type"] == "bearer"


def test_login_wrong_password(client, registered_user):
    """Login fails with an incorrect password."""
    response = client.post("/auth/login", json={
        "email": registered_user["email"],
        "password": "wrongpassword"
    })
    assert response.status_code == 401


def test_login_nonexistent_email(client):
    """Login fails if the email does not exist."""
    response = client.post("/auth/login", json={
        "email": "nobody@example.com",
        "password": "password123"
    })
    assert response.status_code == 401


def test_login_account_lockout(client, registered_user):
    """Account is locked after 5 consecutive failed login attempts."""
    for _ in range(5):
        client.post("/auth/login", json={
            "email": registered_user["email"],
            "password": "wrongpassword"
        })

    # 6th attempt should return 423 Locked
    response = client.post("/auth/login", json={
        "email": registered_user["email"],
        "password": "wrongpassword"
    })
    assert response.status_code == 423


def test_login_counter_resets_after_success(client, registered_user):
    """Failed attempt counter resets after a successful login."""
    # Fail 3 times
    for _ in range(3):
        client.post("/auth/login", json={
            "email": registered_user["email"],
            "password": "wrongpassword"
        })

    # Login successfully
    response = client.post("/auth/login", json=registered_user)
    assert response.status_code == 200

    # Fail again — counter should have reset, not lock yet
    response = client.post("/auth/login", json={
        "email": registered_user["email"],
        "password": "wrongpassword"
    })
    assert response.status_code == 401