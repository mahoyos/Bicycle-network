def test_me_success(client, registered_user, auth_tokens):
    """Authenticated user can retrieve their profile."""
    response = client.get("/auth/me", headers={
        "Authorization": f"Bearer {auth_tokens['access_token']}"
    })
    assert response.status_code == 200
    data = response.json()
    assert data["email"] == registered_user["email"]
    assert "id" in data
    assert "is_active" in data
    assert "hashed_password" not in data


def test_me_no_token(client):
    """Request without token returns 403 (no Authorization header)."""
    response = client.get("/auth/me")
    assert response.status_code == 403


def test_me_invalid_token(client):
    """Request with invalid token returns 401."""
    response = client.get("/auth/me", headers={
        "Authorization": "Bearer this.is.invalid"
    })
    assert response.status_code == 401


def test_me_returns_correct_user(client, auth_tokens, registered_user):
    """The /me endpoint returns the correct authenticated user."""
    response = client.get("/auth/me", headers={
        "Authorization": f"Bearer {auth_tokens['access_token']}"
    })
    assert response.json()["email"] == registered_user["email"]