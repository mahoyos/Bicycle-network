from fastapi_mail import FastMail, MessageSchema, ConnectionConfig, MessageType
from app.core.config import settings

conf = ConnectionConfig(
    MAIL_USERNAME=settings.MAIL_USERNAME,
    MAIL_PASSWORD=settings.MAIL_PASSWORD,
    MAIL_FROM=settings.MAIL_FROM,
    MAIL_PORT=settings.MAIL_PORT,
    MAIL_SERVER=settings.MAIL_SERVER,
    MAIL_STARTTLS=settings.MAIL_STARTTLS,
    MAIL_SSL_TLS=settings.MAIL_SSL_TLS,
    USE_CREDENTIALS=True,
    VALIDATE_CERTS=True,
)

fastmail = FastMail(conf)


async def send_password_recovery_email(email: str, token: str) -> None:
    reset_link = f"{settings.FRONTEND_URL}/reset-password?token={token}"

    message = MessageSchema(
        subject="Password Recovery — Red Bicicletas",
        recipients=[email],
        body=f"""
        <html>
        <body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
            <h2 style="color: #333;">Password Recovery</h2>
            <p>You requested a password reset for your Red Bicicletas account.</p>
            <p>Click the button below to set a new password. This link is valid for <strong>1 hour</strong>.</p>
            <a href="{reset_link}"
               style="display: inline-block; padding: 12px 24px; background-color: #2563eb;
                      color: white; text-decoration: none; border-radius: 6px; margin: 16px 0;">
                Reset Password
            </a>
            <p style="color: #666; font-size: 14px;">
                If you did not request a password reset, you can safely ignore this email.
            </p>
            <p style="color: #666; font-size: 14px;">
                If the button does not work, copy and paste this link into your browser:<br>
                <a href="{reset_link}">{reset_link}</a>
            </p>
        </body>
        </html>
        """,
        subtype=MessageType.html,
    )

    await fastmail.send_message(message)