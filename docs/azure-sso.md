# Azure SSO Setup

Local mode runs without sign-in so the demo is easy to review. Azure mode enables Entra ID login with OIDC, PKCE, verified ID tokens, and an AES-GCM encrypted HttpOnly SameSite session cookie.

## Local Status Check

```bash
curl http://localhost:8080/api/auth/session
```

Local demo output should show:

```json
{
  "mode": "dev",
  "required": false,
  "storage": "encrypted HttpOnly SameSite cookie"
}
```

## Azure App Registration

1. Create an Entra ID app registration.
2. Add a web redirect URI:

```text
https://<your-app-host>/auth/callback
```

For local Azure-mode testing through a tunnel, use the tunnel HTTPS callback.

3. Create a client secret.
4. Copy the tenant ID, client ID, and secret.

## Environment Variables

```bash
AUTH_MODE=azure
AUTH_REQUIRED=true
AUTH_REDIRECT_URL=https://<your-app-host>/auth/callback
AZURE_TENANT_ID=<tenant-id>
AZURE_CLIENT_ID=<app-registration-client-id>
AZURE_CLIENT_SECRET=<client-secret>
AUTH_SESSION_KEY=<base64-encoded-32-byte-key>
COOKIE_SECURE=true
```

Generate the session key:

```bash
openssl rand -base64 32
```

## Expected Behavior

- `GET /api/auth/session` reports `mode: azure`.
- `/auth/login` redirects to Microsoft identity platform.
- `/auth/callback` verifies the ID token and stores only encrypted session claims in an HttpOnly cookie.
- In production, keep `COOKIE_SECURE=true` and serve the API through HTTPS.

## Local Caveat

The checked-in Docker Compose file intentionally uses `AUTH_MODE=dev` so a reviewer can run the app without cloud credentials. Azure SSO is enabled by configuration, not by changing code.
