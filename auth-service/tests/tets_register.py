def test_register_success(client):
    """A new user can register with valid email and password."""
    response = client.post("/auth/register", json={
        "email": "newuser@example.com",
        "password": "securepassword123"
    })
    assert response.status_code == 201
    data = response.json()
    assert data["email"] == "newuser@example.com"
    assert "id" in data
    assert "created_at" in data
    assert "hashed_password" not in data


def test_register_duplicate_email(client, registered_user):
    """Registration fails if the email is already registered."""
    response = client.post("/auth/register", json=registered_user)
    assert response.status_code == 409


def test_register_weak_password(client):
    """Registration fails if the password is shorter than 8 characters."""
    response = client.post("/auth/register", json={
        "email": "user@example.com",
        "password": "123"
    })
    assert response.status_code == 422


def test_register_invalid_email(client):
    """Registration fails if the email format is invalid."""
    response = client.post("/auth/register", json={
        "email": "not-an-email",
        "password": "securepassword123"
    })
    assert response.status_code == 422


def test_register_password_too_long(client):
    """Registration fails if the password exceeds 72 characters."""
    response = client.post("/auth/register", json={
        "email": "user@example.com",
        "password": "a" * 73
    })
    assert response.status_code == 422