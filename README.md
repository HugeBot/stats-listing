# Discord Bot Stats Listing CLI

Stats Listing is a CLI script that retrieve Guild Count from Discord APi and post it to configured websites.

Supports:
- top.gg
- discordlist.gg (`NEW`)
- discords.com
- discord.bots.gg
- And more

Download the [latest release](https://github.com/HugeBot/stats-listing/releases/latest).

### Configuration file
To execute this script programatically you need to create a config.yaml file on the script root folder, the config file structure:
```yaml
botToken: XXXXXX

websites:
  - name: Top.GG
    apiPath: https://top.gg/api/bots/@bot_id@/stats
    token: XXXXXX
    bodyPattern: '{"guild_count": @guild_count@, "shard_count": @guild_count@}'
  - name: Discords.Bots.GG
    apiPath: https://discord.bots.gg/api/v1/bots/@bot_id@/stats
    token: XXXXXX
    bodyPattern: '{"guildCount": @guild_count@}'
  - name: Discords.com
    apiPath: https://discords.com/bots/api/bot/@bot_id@
    token: XXXXXX
  - name: Dlist.GG
    apiPath: https://api.discordlist.gg/v0/bots/@bot_id@/guilds?count=@guild_count@
    token: 'Bearer XXXXXX'
    method: PUT
```

Automatic variables for the config file:
- ``@bot_id@`` replaces this with the bot id when the scripts init.
- ``@guild_count@`` replaces this with the GuildCount.