# Breach Lookup

Breach Lookup is a cli tool to lookup domains in breach databases via breach.vip API.

https://breach.vip/

https://breach.vip/api/docs

## Usage

```bash
Usage of BreachLookup:
  -case
        Case sensitive search (optional)
  -categories string
        Comma-separated categories (optional)
  -fields string
        Comma-separated fields (required) (default "domain")
  -out string
        Output file path (default "output.json")
  -term string
        Search term (required)
  -url string
        API endpoint URL (default "https://breach.vip/api/search")
  -wildcard
        Enable wildcard (optional)
```

```bash
BreachLookup -term 'example.com'
Searching for example.com on https://breach.vip/api/search
Response saved to output.json
Emails saved to emails_output.json
Passwords saved to passwords_output.json
Emails: 3311
Passwords: 1329
Total results: 3311 (Maximum: 10000)
```

## Todo

- [ ] Add support to lookup hashes against free hash databases
- [ ] Change password output to only show cleartext passwords
- [ ] CSV output