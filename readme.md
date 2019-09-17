# Slack Slash Command to get Specific Timezone's Current Time

- Related Blog Post link (in Korean): https://jusths.tistory.com/151

## Implementation Phase 1. Local Server which parses Slack's HTTP Post

1. Set environment variables (Windows)

  ```$set SLACK_SIGNING_SECRET=<Slack App's Signing Secret>```

2. Locally, run the server

  ```$go run main.go```

3. Port Forwarding, so Slack HTTP POST can reach the server

  ```$ssh -o ServerAliveInterval=60 -R 80:localhost:8080 serveo.net```