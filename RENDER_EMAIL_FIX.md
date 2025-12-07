# 🚀 Render Deployment Fix Guide

## 🔴 Problem
Render blocks outbound connections on port 587 (STARTTLS), causing:
```
dial tcp 172.253.132.109:587: connect: connection timed out
```

## ✅ Solution
Use port **465** (SSL/TLS) instead, which Render allows.

---

## 📋 Steps to Fix on Render

### 1. Update Environment Variable

Go to your Render dashboard:
1. Navigate to your service: **car-wash-app-j54r**
2. Click **Environment** tab
3. Find `SMTP_PORT` variable
4. Change value from `587` to `465`
5. Click **Save Changes**

### 2. Verify Other SMTP Variables

Make sure these are set correctly:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=465                          ← Changed from 587
SMTP_USERNAME=ojoolabanji59@gmail.com
SMTP_PASSWORD=[YOUR_APP_PASSWORD]      ← Must be 16-char App Password!
FROM_EMAIL=ojoolabanji59@gmail.com
FROM_NAME=Banji CarWash App
```

### 3. Get Gmail App Password (If Not Done)

1. Go to https://myaccount.google.com/security
2. Enable **2-Step Verification** (if not enabled)
3. Go to https://myaccount.google.com/apppasswords
4. Select **Mail** and **Windows Computer**
5. Click **Generate**
6. Copy the 16-character password (e.g., `abcd efgh ijkl mnop`)
7. Remove spaces: `abcdefghijklmnop`
8. Paste into Render's `SMTP_PASSWORD` variable

### 4. Deploy Updated Code

Your code now supports both ports automatically:
- Port **587**: Uses STARTTLS (for local dev)
- Port **465**: Uses SSL/TLS (for Render)

**Deploy options:**

**Option A: Manual Deploy**
1. Go to Render dashboard
2. Click **Manual Deploy** → **Deploy latest commit**

**Option B: Git Push (if auto-deploy enabled)**
```bash
git add .
git commit -m "Fix: Support port 465 for email on Render"
git push origin main
```

### 5. Test

After deployment:
1. Register a new user on your frontend
2. Check email inbox for verification code
3. Verify the email works

---

## 🔍 How the Fix Works

### Before (Port 587 only):
```go
// Only worked with STARTTLS
smtp.SendMail(host+":587", auth, from, to, msg)
```
❌ Render blocks this

### After (Both ports supported):
```go
if port == "465" {
    // Use SSL/TLS (works on Render)
    conn := tls.Dial("tcp", host+":465", tlsConfig)
    // ... send via SSL
} else {
    // Use STARTTLS (works locally)
    smtp.SendMail(host+":587", auth, from, to, msg)
}
```
✅ Works everywhere!

---

## 📊 Port Comparison

| Port | Protocol | Local Dev | Render | Gmail |
|------|----------|-----------|--------|-------|
| 587  | STARTTLS | ✅ Works  | ❌ Blocked | ✅ Supported |
| 465  | SSL/TLS  | ✅ Works  | ✅ Works | ✅ Supported |
| 25   | Plain    | ⚠️ Risky  | ❌ Blocked | ❌ Not supported |

**Recommendation:** Use **465** for production (Render, Heroku, etc.)

---

## ✅ Checklist

- [ ] Changed `SMTP_PORT` to `465` on Render
- [ ] Set `SMTP_PASSWORD` to Gmail App Password (16 chars)
- [ ] Deployed updated code to Render
- [ ] Tested registration → email received
- [ ] Verified email verification works

---

## 🐛 Troubleshooting

### Still getting timeout?
- Double-check `SMTP_PORT=465` (not 587)
- Verify App Password is correct (no spaces)
- Check Render logs for new error messages

### Authentication failed?
- Make sure you're using **App Password**, not regular password
- App Password must be 16 characters (no spaces)
- Generate new App Password if unsure

### Email goes to spam?
- Normal for first few emails
- Ask users to check spam folder
- Mark as "Not Spam" to train Gmail

---

## 📝 Local Development

Your local `.env` can still use port 587:
```env
SMTP_PORT=587  # Works fine locally
```

The code automatically handles both ports! 🎉
