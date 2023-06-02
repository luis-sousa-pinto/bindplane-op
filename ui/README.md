# BindPlane UI

This directory contains the UI portion for BindPlane which is a single page react app.

## VS Code

Recommended Plugins:

- `esbenp.prettier-vscode`

Settings:

```json
{
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.formatOnSave": true
}
```

## GraphQL Subscriptions on Chrome

Its a known issue that the websocket connections needed for live updates for agents and configurations does not work for Chrome. It is known to work on Safari and Firefox.
