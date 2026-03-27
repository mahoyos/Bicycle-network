def test_create_and_read_event(client):
    payload = {
        "name": "Test Event",
        "type": "Tour",
        "date": "2026-12-01T10:00:00",
        "description": "Test description",
        "start_location_lat": 40.7128,
        "start_location_lng": -74.0060,
        "end_location_lat": 34.0522,
        "end_location_lng": -118.2437
    }

    create_resp = client.post("/events/", json=payload)
    assert create_resp.status_code == 200
    event = create_resp.json()
    assert event["id"] > 0
    assert event["name"] == payload["name"]
    assert event["type"] == payload["type"]
    assert event["start_location_lat"] == payload["start_location_lat"]

    read_resp = client.get(f"/events/{event['id']}")
    assert read_resp.status_code == 200
    read_event = read_resp.json()
    assert read_event["name"] == payload["name"]
    assert read_event["end_location_lng"] == payload["end_location_lng"]


def test_update_event(client):
    payload = {
        "name": "Updatable Event",
        "type": "Route",
        "date": "2026-12-10T08:00:00",
        "description": "Update me",
        "start_location_lat": 10.0,
        "start_location_lng": 20.0,
        "end_location_lat": 30.0,
        "end_location_lng": 40.0
    }
    create_resp = client.post("/events/", json=payload)
    assert create_resp.status_code == 200
    event_id = create_resp.json()["id"]

    update_payload = {
        "name": "Updated Name",
        "type": "Competition",
        "date": "2026-12-10T08:00:00",
        "description": "Updated description",
        "start_location_lat": 11.0,
        "start_location_lng": 21.0,
        "end_location_lat": 31.0,
        "end_location_lng": 41.0
    }
    update_resp = client.put(f"/events/{event_id}", json=update_payload)
    assert update_resp.status_code == 200
    updated = update_resp.json()

    assert updated["name"] == "Updated Name"
    assert updated["type"] == "Competition"
    assert updated["start_location_lat"] == update_payload["start_location_lat"]


def test_delete_event(client):
    payload = {
        "name": "Delete Event",
        "type": "Route",
        "date": "2026-09-01T07:00:00",
        "description": "To be removed",
        "start_location_lat": 1.0,
        "start_location_lng": 2.0,
        "end_location_lat": 3.0,
        "end_location_lng": 4.0
    }
    create_resp = client.post("/events/", json=payload)
    event_id = create_resp.json()["id"]

    delete_resp = client.delete(f"/events/{event_id}")
    assert delete_resp.status_code == 200
    assert delete_resp.json()["message"] == "Event deleted successfully"

    missing_resp = client.get(f"/events/{event_id}")
    assert missing_resp.status_code == 404


def test_registrations_crud_flow(client):
    event_payload = {
        "name": "Registration Event",
        "type": "Competition",
        "date": "2026-10-10T09:30:00",
        "description": "Event with registration",
        "start_location_lat": 50.0,
        "start_location_lng": 60.0,
        "end_location_lat": 70.0,
        "end_location_lng": 80.0
    }
    event = client.post("/events/", json=event_payload).json()
    event_id = event["id"]

    reg_resp = client.post(
        f"/events/{event_id}/registrations", json={"user_id": 42})
    assert reg_resp.status_code == 200
    reg_data = reg_resp.json()
    assert reg_data["user_id"] == 42

    list_resp = client.get("/users/42/registrations")
    assert list_resp.status_code == 200
    assert any(item["event_id"] == event_id for item in list_resp.json())

    delete_reg_resp = client.delete(f"/events/{event_id}/registrations/42")
    assert delete_reg_resp.status_code == 200
    assert delete_reg_resp.json(
    )["message"] == "Successfully unregistered from event"


def test_registration_on_nonexistent_event_returns_404(client):
    bad_resp = client.post("/events/9999/registrations", json={"user_id": 100})
    assert bad_resp.status_code == 404
