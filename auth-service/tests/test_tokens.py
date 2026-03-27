def test_refresh_token_success(client, auth_tokens):
    """A valid refresh token returns a new access token."""
    assert "refresh_token" in auth_tokens, "Login failed, no tokens available"
    response = client.post("/auth/refresh", json={
        "refresh_token": auth_tokens["refresh_token"]
    })
    assert response.status_code == 200
    data = response.json()
    assert "access_token" in data


def test_refresh_token_invalid(client):
    """An invalid refresh token returns 401."""
    response = client.post("/auth/refresh", json={
        "refresh_token": "this.is.not.valid"
    })
    assert response.status_code == 401


def test_logout_revokes_refresh_token(client, auth_tokens):
    """After logout, the refresh token cannot be used again."""
    assert "refresh_token" in auth_tokens, "Login failed, no tokens available"

    client.post("/auth/logout", json={
        "refresh_token": auth_tokens["refresh_token"]
    })

    response = client.post("/auth/refresh", json={
        "refresh_token": auth_tokens["refresh_token"]
    })
    assert response.status_code == 401


def test_logout_success(client, auth_tokens):
    """Logout returns 200 with a confirmation message."""
    assert "refresh_token" in auth_tokens, "Login failed, no tokens available"
    response = client.post("/auth/logout", json={
        "refresh_token": auth_tokens["refresh_token"]
    })
    assert response.status_code == 200
    assert "message" in response.json()