#!/usr/bin/env python3

import argparse
import concurrent.futures
import http.cookiejar
import json
import re
import time
import urllib.error
import urllib.request
import uuid


class Client:
    def __init__(self, api_base: str):
        self.api_base = api_base.rstrip("/")
        self.cookies = http.cookiejar.CookieJar()
        self.opener = urllib.request.build_opener(
            urllib.request.HTTPCookieProcessor(self.cookies)
        )

    def request(self, method: str, path: str, data: dict | None = None) -> tuple[int, dict]:
        body = None if data is None else json.dumps(data).encode()
        request = urllib.request.Request(
            f"{self.api_base}{path}",
            data=body,
            method=method,
            headers={"Content-Type": "application/json"},
        )
        try:
            with self.opener.open(request, timeout=10) as response:
                return response.status, json.load(response)
        except urllib.error.HTTPError as error:
            return error.code, json.load(error)


def mailpit_request(mailpit_base: str, method: str, path: str) -> dict:
    request = urllib.request.Request(
        f"{mailpit_base.rstrip('/')}{path}",
        method=method,
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        content = response.read()
        if not content or not response.headers.get_content_type() == "application/json":
            return {}
        return json.loads(content)


def clear_mailpit(mailpit_base: str) -> None:
    mailpit_request(mailpit_base, "DELETE", "/api/v1/messages")


def wait_for_message(mailpit_base: str, recipient: str, subject: str) -> dict:
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        messages = mailpit_request(mailpit_base, "GET", "/api/v1/messages")
        for summary in messages.get("messages", []):
            recipients = [address.get("Address", "") for address in summary.get("To", [])]
            if recipient in recipients and summary.get("Subject") == subject:
                return mailpit_request(
                    mailpit_base,
                    "GET",
                    f"/api/v1/message/{summary['ID']}",
                )
        time.sleep(0.2)
    raise RuntimeError(f"Mailpit did not receive {subject!r} for {recipient}")


def extract_token(message: dict, route: str) -> str:
    content = f"{message.get('HTML', '')}\n{message.get('Text', '')}"
    match = re.search(rf"{re.escape(route)}\?token=([0-9a-f]{{64}})", content)
    if not match:
        raise RuntimeError(f"email did not contain a {route} token")
    return match.group(1)


def expect(status: int, body: dict, expected_status: int, expected_code: str | None = None) -> None:
    if status != expected_status:
        raise RuntimeError(f"HTTP {status}, expected {expected_status}: {body}")
    if expected_code is not None and body.get("error", {}).get("code") != expected_code:
        raise RuntimeError(f"error code {body.get('error', {}).get('code')!r}, expected {expected_code!r}")


def verify(api_base: str, mailpit_base: str) -> None:
    clear_mailpit(mailpit_base)
    suffix = uuid.uuid4().hex[:12]
    email = f"auth-{suffix}@example.local"
    username = f"auth{suffix}"
    old_password = "SecurePass123!"
    new_password = "NewSecurePass456!"
    client = Client(api_base)

    status, body = client.request(
        "POST",
        "/auth/register",
        {"email": email, "username": username, "password": old_password},
    )
    expect(status, body, 201)
    status, body = client.request(
        "POST",
        "/auth/register",
        {"email": email, "username": f"other{suffix}", "password": old_password},
    )
    expect(status, body, 409, "EMAIL_ALREADY_EXISTS")
    status, body = client.request(
        "POST",
        "/auth/register",
        {"email": f"other-{suffix}@example.local", "username": username, "password": old_password},
    )
    expect(status, body, 409, "USERNAME_ALREADY_EXISTS")
    print("PASS registration error codes match the frontend contract")

    verification_message = wait_for_message(
        mailpit_base,
        email,
        "Verify your email - ACM Hot 100",
    )
    verification_token = extract_token(verification_message, "verify-email")
    print("PASS registration email captured by Mailpit")

    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": "WrongPass123!"}
    )
    expect(status, body, 401, "INVALID_CREDENTIALS")
    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": old_password}
    )
    expect(status, body, 403, "EMAIL_NOT_VERIFIED")
    print("PASS pending-account login does not leak on wrong password")

    def verify_once() -> tuple[int, dict]:
        return Client(api_base).request(
            "POST", "/auth/verify-email", {"token": verification_token}
        )

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        verification_results = list(executor.map(lambda _: verify_once(), range(2)))
    verification_statuses = sorted(status for status, _ in verification_results)
    if verification_statuses != [200, 400]:
        raise RuntimeError(f"concurrent verification statuses: {verification_statuses}")
    print("PASS verification token consumed exactly once")

    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": old_password}
    )
    expect(status, body, 200)
    status, body = client.request("POST", "/auth/refresh")
    expect(status, body, 200)
    status, body = client.request("POST", "/auth/refresh")
    expect(status, body, 200)
    print("PASS login and two refresh rotations")

    session_before_logout = Client(api_base)
    for cookie in client.cookies:
        session_before_logout.cookies.set_cookie(cookie)

    status, body = client.request("POST", "/auth/logout")
    expect(status, body, 200)
    status, body = session_before_logout.request("POST", "/auth/refresh")
    expect(status, body, 401)
    print("PASS logout invalidates refresh token")

    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": old_password}
    )
    expect(status, body, 200)
    session_before_reset = Client(api_base)
    for cookie in client.cookies:
        session_before_reset.cookies.set_cookie(cookie)

    status, body = client.request(
        "POST", "/auth/forgot-password", {"email": email}
    )
    expect(status, body, 200)
    reset_message = wait_for_message(
        mailpit_base,
        email,
        "Reset your password - ACM Hot 100",
    )
    reset_token = extract_token(reset_message, "reset-password")
    print("PASS password reset email captured by Mailpit")

    def reset_once() -> tuple[int, dict]:
        return Client(api_base).request(
            "POST",
            "/auth/reset-password",
            {"token": reset_token, "new_password": new_password},
        )

    with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
        reset_results = list(executor.map(lambda _: reset_once(), range(2)))
    reset_statuses = sorted(status for status, _ in reset_results)
    if reset_statuses != [200, 400]:
        raise RuntimeError(f"concurrent reset statuses: {reset_statuses}")
    status, body = session_before_reset.request("POST", "/auth/refresh")
    expect(status, body, 401)
    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": old_password}
    )
    expect(status, body, 401, "INVALID_CREDENTIALS")
    status, body = client.request(
        "POST", "/auth/login", {"email": email, "password": new_password}
    )
    expect(status, body, 200)
    print("PASS reset token consumed once, old session/password revoked")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Verify the complete auth and Mailpit flow")
    parser.add_argument("--api", default="http://127.0.0.1:18080/api/v1")
    parser.add_argument("--mailpit", default="http://127.0.0.1:18025")
    args = parser.parse_args()
    verify(args.api, args.mailpit)
