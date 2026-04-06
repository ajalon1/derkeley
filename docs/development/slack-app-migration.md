# Slack App Migration Guide

## Overview

This guide provides step-by-step instructions for migrating the DataRobot CLI GitHub Actions from deprecated Slack Incoming Webhooks to a native Slack App with OAuth token authentication.

**Why Migrate?**
- Incoming Webhooks are a deprecated custom integration with limited future support
- Native Slack Apps provide better security, audit logging, and permission controls
- OAuth tokens enable more granular permission scoping
- Ensures long-term compatibility as Slack evolves

---

## Prerequisites

- Access to the DataRobot Slack workspace as an admin or app manager
- Access to GitHub repository settings with admin permissions
- Access to 1Password "DataRobot Integrations" vault
- Basic understanding of Slack App management
- ~30-45 minutes to complete the migration

---

## Step 1: Create a Native Slack App

### 1.1 Access Slack App Management

1. Go to [Slack API Apps Dashboard](https://api.slack.com/apps)
2. Sign in with your Slack workspace account
3. Click **Create New App**
4. Choose **From scratch** (not "From an app manifest")
5. Fill in the form:
   - **App name:** `DataRobot CLI Notifications`
   - **Workspace:** Select your DataRobot Slack workspace
6. Click **Create App**

### 1.2 Configure Basic Information

1. On the app settings page, go to **Basic Information** (left sidebar)
2. Under **App-Level Tokens**, click **Generate Token and Scopes**
3. Name the token: `github-actions-token`
4. Add scopes: `chat:write` (required for posting messages)
5. Click **Generate** and copy the token (format: `xapp-1-...`)
6. **Store this token immediately in 1Password:**
   - Vault: "DataRobot Integrations"
   - Item name: "Slack App - DataRobot CLI Notifications (OAuth)"
   - Include fields: creation_date, scope, app_id, workspace_name
7. Click **Done**

### 1.3 Enable Bot Token Permissions

1. Go to **OAuth & Permissions** (left sidebar)
2. Under **Scopes**, go to **Bot Token Scopes**
3. Click **Add an OAuth Scope**
4. Add the following scopes:
   - `chat:write` - Post messages in channels
   - `chat:write.public` - Post in public channels (if needed)
   - `users:read` - Read user information (optional, for better logging)
5. Scroll to top and note your **Bot User OAuth Token** (format: `xoxb-...`)
6. **Store in 1Password** in the same item from Step 1.2

### 1.4 Install App to Workspace

1. At the top of **OAuth & Permissions**, click **Install to Workspace**
2. Review the permissions and click **Allow**
3. You're redirected back to your app settings
4. Confirm the Bot User OAuth Token is visible
5. Take note of your **App ID** (visible on **Basic Information** page)

---

## Step 2: Configure the Slack App

### 2.1 Set Up Event Subscriptions (Optional but Recommended)

If you want to enable more advanced features in the future:

1. Go to **Event Subscriptions** (left sidebar)
2. Toggle **Enable Events** to ON
3. For Request URL, you can skip this for now (not needed for posting messages)
4. Under **Subscribe to bot events**, you don't need to add events for basic notification posting

### 2.2 Add App to Target Channels

1. In your DataRobot Slack workspace, go to the channels where you want notifications:
   - `#releases` (for release notifications)
   - `#devops` (for smoke test failures)
   - Any other relevant channels

2. In each channel:
   - Click the channel name at the top
   - Go to **Integrations** tab
   - Click **Add an App**
   - Search for and select **DataRobot CLI Notifications**
   - Confirm addition

---

## Step 3: Update GitHub Actions Workflows

### 3.1 Identify Current Webhook Usage

The current workflows using `SLACK_WEBHOOK_URL`:
- `.github/workflows/release.yaml` - Release notifications
- `.github/workflows/smoke-tests.yaml` - Smoke test failure alerts
- `.github/workflows/smoke-tests-on-demand.yaml` - On-demand test alerts

### 3.2 Choose Your Notification Method

#### Option A: Use slack-notify Action (Recommended)

This is the simplest approach using a third-party GitHub Action.

1. In each affected workflow, find the Slack notification step
2. Replace the current notification step with:

```yaml
- name: Notify Slack on Success
  if: success()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {
        "text": "DR CLI Release Successful",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*DataRobot CLI Release* ✅\n*Version:* ${{ github.ref_name }}\n*Status:* Success"
            }
          }
        ]
      }
  env:
    SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
    SLACK_WEBHOOK_TYPE: INCOMING_WEBHOOK
```

Wait - use the new app's Bot Token instead:

```yaml
- name: Notify Slack on Success
  if: success()
  uses: slackapi/slack-github-action@v1
  with:
    channel: '#releases'
    payload: |
      {
        "text": "DR CLI Release Successful",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*DataRobot CLI Release* ✅\n*Version:* ${{ github.ref_name }}\n*Status:* Success"
            }
          }
        ]
      }
  env:
    SLACK_BOT_TOKEN: ${{ secrets.SLACK_BOT_TOKEN }}
```

3. Update the GitHub Actions secret:
   - Old secret: `SLACK_WEBHOOK_URL` (keep for now during testing)
   - New secret: `SLACK_BOT_TOKEN` (create new)
   - Value: The Bot User OAuth Token from Step 1.4 (format: `xoxb-...`)

#### Option B: Direct API Call (More Control)

If you prefer full control without third-party actions:

```yaml
- name: Notify Slack on Failure
  if: failure()
  run: |
    curl -X POST \
      -H 'Authorization: Bearer ${{ secrets.SLACK_BOT_TOKEN }}' \
      -H 'Content-Type: application/json' \
      -d '{
        "channel": "#devops",
        "text": "DR CLI Smoke Tests Failed",
        "blocks": [{
          "type": "section",
          "text": {
            "type": "mrkdwn",
            "text": "*Smoke Tests Failed* ❌\n*Workflow:* ${{ github.workflow }}\n*Branch:* ${{ github.ref_name }}"
          }
        }]
      }' \
      https://slack.com/api/chat.postMessage
```

### 3.3 Update GitHub Secrets

1. Go to repository **Settings** → **Secrets and variables** → **Actions**
2. Create new secret:
   - Name: `SLACK_BOT_TOKEN`
   - Value: Bot User OAuth Token from Step 1.4 (format: `xoxb-...`)
   - Click **Save**
3. Keep the old `SLACK_WEBHOOK_URL` secret for now (for fallback during transition)

---

## Step 4: Test the Migration

### 4.1 Test on a Single Workflow First

1. Trigger a smoke test manually:
   - Go to **Actions** → **Daily Smoke Tests** → **Run workflow**
2. Monitor the workflow run
3. Check that the Slack notification appears in the target channel
4. Verify the message format and content are correct

### 4.2 Verify Bot User in Channel

In Slack, check that the bot user appears:
- Go to `#releases` or `#devops` channel
- Click channel name → **Members** tab
- Verify **DataRobot CLI Notifications** bot is listed

### 4.3 Check Workflow Logs

1. In GitHub Actions, view the workflow run logs
2. Look for the Slack notification step
3. Verify no errors about authentication or channel access
4. If using curl, check the response for success status

---

## Step 5: Complete the Migration

### 5.1 Update All Workflows

Once testing is successful:

1. Update all workflows that use `SLACK_WEBHOOK_URL`:
   - `.github/workflows/release.yaml`
   - `.github/workflows/smoke-tests.yaml`
   - `.github/workflows/smoke-tests-on-demand.yaml`
   - Any other workflows using Slack notifications

2. Replace with new `SLACK_BOT_TOKEN` and chosen notification method (Option A or B)

### 5.2 Final Testing

1. Trigger each workflow type:
   - Release workflow (if applicable)
   - Smoke tests
   - Any manual triggers
2. Verify notifications appear in correct channels
3. Monitor for 24 hours to ensure stability

### 5.3 Clean Up Old Integration

1. **Revoke the old webhook:**
   - Go to Slack workspace
   - Navigate to **[DataRobot Slack](https://datarobot.slack.com/apps/manage/custom-integrations)** → **Manage Apps** → **Custom Integrations**
   - Find the old "Incoming Webhooks" integration
   - Click and select the specific webhook for DR CLI
   - Click **Delete** or **Revoke**

2. **Archive old secret in 1Password:**
   - Go to "SLACK_WEBHOOK_URL" item in 1Password
   - Add note: "Migrated to native Slack App on [DATE]"
   - Move to archive folder (do not delete - keep for audit trail)

3. **Optionally delete GitHub secret:**
   - Go to repository **Settings** → **Secrets and variables** → **Actions**
   - Find `SLACK_WEBHOOK_URL`
   - Click **Delete** (only after confirming all workflows updated)

---

## Troubleshooting

### Bot Not Appearing in Channel

**Problem:** Slack notification fails with "channel_not_found" error.

**Solution:**
1. Ensure the bot is invited to the channel
2. In target channel, go to **Integrations** → **Add an App**
3. Search for "DataRobot CLI Notifications" and add it
4. Try the workflow again

### Authentication Errors

**Problem:** Workflow fails with "invalid_token" or "token_revoked" error.

**Solution:**
1. Verify `SLACK_BOT_TOKEN` secret is set correctly in GitHub
2. Check token format (should start with `xoxb-`)
3. Regenerate token in Slack App dashboard if needed:
   - Go to **OAuth & Permissions**
   - Copy the **Bot User OAuth Token** again
   - Update GitHub secret
4. Ensure token hasn't been manually revoked in Slack

### Message Not Formatting Correctly

**Problem:** Slack message appears but formatting is wrong (blocks not rendering, etc.).

**Solution:**
1. Validate JSON in the workflow payload
2. Test with simpler message format first:
   ```json
   {"text": "Simple test message"}
   ```
3. Gradually add formatting (blocks, markdown) once basic messages work
4. Refer to [Slack Block Kit Builder](https://app.slack.com/block-kit) to design messages

### Rate Limiting Issues

**Problem:** Workflow fails with "rate_limited" error.

**Solution:**
1. Slack allows up to 1 message per second per workspace
2. If workflows are sending multiple messages, stagger them
3. Combine multiple notifications into a single message using blocks
4. Implement retry logic with exponential backoff in workflows:
   ```yaml
   - name: Send Slack Notification
     uses: slackapi/slack-github-action@v1
     with:
       retry-max-times: 3
       retry-wait: 5
   ```

### Wrong Channel Receiving Notifications

**Problem:** Notifications go to wrong Slack channel.

**Solution:**
1. In workflow, verify the `channel` parameter is correct
2. Channel format should be `#channel-name` or channel ID
3. Check that the bot is invited to the intended channel
4. If using webhook URL style, verify the webhook URL points to correct channel
5. Test by posting manually to verify bot access

---

## Verification Checklist

After completing migration, verify:

- [ ] New Slack App created in Slack workspace
- [ ] Bot User OAuth Token generated and stored
- [ ] Token stored securely in 1Password
- [ ] GitHub secret `SLACK_BOT_TOKEN` created with token value
- [ ] Bot added to all target channels (#releases, #devops, etc.)
- [ ] All workflows updated to use new authentication method
- [ ] Test workflows run successfully and send notifications
- [ ] Notifications appear in correct channels with proper formatting
- [ ] Old webhook revoked in Slack
- [ ] Old secret archived in 1Password
- [ ] Old GitHub secret deleted (optional, after confirmation period)
- [ ] Team members notified of changes
- [ ] Documentation updated in project README/docs

---

## Rollback Plan

If issues occur after migration:

1. **Keep the old `SLACK_WEBHOOK_URL` secret temporarily** (don't delete immediately)
2. **Revert workflow changes** to use the old webhook
3. **Investigate the issue** with new app setup
4. **Test thoroughly** before attempting migration again
5. **Document** what went wrong for future reference

---

## Best Practices Going Forward

1. **Token Rotation:**
   - Rotate `SLACK_BOT_TOKEN` every 90 days
   - Generate new token in Slack App → **OAuth & Permissions**
   - Update GitHub secret
   - Revoke old token in Slack dashboard

2. **Monitoring:**
   - Monitor Slack API logs for failed message posts
   - Set up alerts if notifications stop working
   - Regularly verify bot presence in channels

3. **Permissions:**
   - Regularly audit bot token scopes
   - Remove unused scopes to minimize security exposure
   - Use fine-grained permissions instead of wildcard scopes

4. **Audit Trail:**
   - Keep old webhook details in 1Password for compliance
   - Document migration date and reason
   - Track any issues or rollbacks

---

## Related Documentation

- [Slack API Documentation](https://api.slack.com)
- [Slack Bot User OAuth Tokens](https://api.slack.com/authentication/token-types#bot)
- [Slack Block Kit](https://api.slack.com/block-kit)
- [GitHub Actions Slack Integration](https://github.com/marketplace/actions/slack-notify)
- [DataRobot CLI Secrets Rotation Guide](./secrets-rotation.md)

---

## Questions or Issues?

If you encounter issues during migration:

1. Check the **Troubleshooting** section above
2. Review [Slack API error codes](https://api.slack.com/methods/chat.postMessage#errors)
3. Contact the DevOps or infrastructure team
4. Check GitHub Actions workflow logs for detailed error messages

---

## Version History

| Date | Changes | Author  |
|------|---------|--------|
| 2026-03-26 | Initial migration guide | AI Agent |
