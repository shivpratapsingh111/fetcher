# fetcher
It is made to fetch content of text based files from urls passed to it and save locally.
Purposely for save js files.

## Installation

```
go install github.com/shivpratapsingh111/fetcher@latest
```

### OR

## Build

```
git clone https://github.com/shivpratapsingh111/fetcher && cd fetcher && go build && cp fetcher /usr/local/bin
```

## Usage

- ```fetcher -f urls.txt``` 
> Provide a file containing urls one per line and it will save them locally in `fetched/` directory.

## Options


`-f filepath`
> File containing URLs (one per line)

Example: ```fetcher -f urls.txt -H "Cookie:session=abcd,X-Token:my-x-token"```

---

`-dir directory`
> It will create directory to save output in (default "fetched")
  
Example: ```fetcher -f urls.txt -dir ~/Downloads/myFetchedFiles```

---

`-H "headers"`
> Headers to send with request (comma-separated)

Example: ```fetcher -f urls.txt -H "Cookie:session=abcd,X-Token:my-x-token"```

---

`-proxy IP:PORT` **NOT WORKING CURRENTLY**
> Send all request through proxy (format: IP:PORT)

Example: ```fetcher -f urls.txt -proxy 127.0.0.1:8080```

---

`-r int`
> Number of retries, if a url doesn't responds for the first time (default 3)

Example: ```fetcher -f urls.txt -r 6```

---

`-ra`
> Use a random user-agent for each request

Example: ```fetcher -f urls.txt -ra```

---

`-silent`
> Silent mode, only output URLs that are fetched successfully

Example: ```fetcher -f urls.txt -silent```

---

`-t int`
Number of threads to use (default 60)

Example: ```fetcher -f urls.txt -t 100```

---

`-x int`
> Timeout: Number of seconds to wait for response in seconds (default 12)

Example: ```fetcher -f urls.txt -x 5```

---

## TODO

- [ ] Fix `-proxy` flag, when used gives error `[x] Failed to read response body for [$urls] after [$retires] attempts [tcp: user cancelled]`

- [ ] Improve code quality

---

## Open for contribuitons