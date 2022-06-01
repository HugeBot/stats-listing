# Discord Bot Stats Listing CLI

Stats Listing is a CLI script that retrieve Guild Count from a redis HMap called **shard-stats** and post it to configured websites.

Supports:
- top.gg
- discordlist.gg (`NEW`)
- discords.com
- discord.bots.gg
- And more

Download the [latest release](https://github.com/HugeBot/stats-listing/releases/latest).

### `SHARD-STATS` Redis HMAP structure
* * *
The HMap requires to have the name **shard-stats** and follow this JSON structure:
```json
{
  "id": 0,
  "status": "CONNECTED",          // Not required
  "guildsCacheSize": 1000,
  "usersCacheSize": 1000,         // Not required
  "updatedAt": 1234567891234567   // Not required
}
```
HMap HGETALL Representation
```
redis-cli > HGETALL shard-stats
```
```
1) 0
2) "{\"id\":0,\"status\":\"CONNECTED\",\"guildsCacheSize\":1150,\"usersCacheSize\":0,\"updatedAt\":1234567891234567}"
3) 1
4) "{\"id\":1,\"status\":\"CONNECTED\",\"guildsCacheSize\":1200,\"usersCacheSize\":0,\"updatedAt\":1234567891234567}"
5) 2
6) "{\"id\":2,\"status\":\"CONNECTED\",\"guildsCacheSize\":1400,\"usersCacheSize\":0,\"updatedAt\":1234567891234567}"
```

### Configuration file
To execute this script programatically you need to create a config.yaml file on the script root folder, the config file structure:
```yaml
botId: XXXXXX

redis:
  host: localhost
  pass: password
  port: 6379
  db: 0

websites:
  - name: Top.GG
    apiPath: https://top.gg/api/bots/@bot_id@/stats
    token: XXXXXX
    bodyPattern: '{"server_count": @server_count@, "shard_count": @shard_count@}'
  - name: Discords.Bots.GG
    apiPath: https://discord.bots.gg/api/v1/bots/@bot_id@/stats
    token: XXXXXX
    bodyPattern: '{"guildCount": @server_count@}'
  - name: Discords.com
    apiPath: https://discords.com/bots/api/bot/@bot_id@
    token: XXXXXX
  - name: Dlist.GG
    apiPath: https://api.discordlist.gg/v0/bots/@bot_id@/guilds?count=@server_count@
    token: 'Bearer XXXXXX'
    method: PUT
```

Automatic variables for the config file:
- ``@bot_id@`` replaces this with the bot id when the scripts init.
- ``@server_count@`` replaces this with the GuildCount.
- ``@shard_count@`` replaces this with the ShardCount.

HMap Json Representation
```json
[
  {
    "id": 0,
    "status": "CONNECTED",
    "guildsCacheSize": 1150,
    "usersCacheSize": 0,
    "updatedAt": 1234567891234567
  },
  {
    "id": 1,
    "status": "CONNECTED",
    "guildsCacheSize": 1200,
    "usersCacheSize": 0,
    "updatedAt": 1234567891234567
  },
  {
    "id": 2,
    "status": "CONNECTED",
    "guildsCacheSize": 1400,
    "usersCacheSize": 0,
    "updatedAt": 1234567891234567
  }
]
```

The script will fetch all values from the HMap with HGETALL and compile into a simple structure:
```json
{
  "ServerCount": 3750,
  "ShardCount": 3
}
```
