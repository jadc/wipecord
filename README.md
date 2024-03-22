# wipecord
Deletes all messages from yourself (or another user) in a Discord guild

## Usage
Delete all messages from yourself in a guild
```bash
./wipecord -t "DISCORD_TOKEN" -g "GUILD_ID"
```
```bash
DISCORD_TOKEN="DISCORD_TOKEN" ./wipecord -g "GUILD_ID"
```

Delete all messages from another user in a guild (requires you have manage messages permission)
```bash
./wipecord -t "DISCORD_TOKEN" -g "GUILD_ID" -a "AUTHOR_ID"
```

### Flags
```bash
  -t string
    	Discord Token
  -g string
    	Guild ID
  -a string
    	Author ID (default "@me")
  -f int
    	Delay between message fetches in milliseconds (default 30000)
  -r int
    	Delay between message deletions in milliseconds (default 1500)
  -y	Skip confirmations
```

## Todo
- [ ] Add better input validation
- [ ] Create tests
- [ ] Create GitHub action to continuously test and build into a GitHub release
