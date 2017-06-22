README
========

Command-line utility for the UrbanAirship Reach API. Currently, it is not possible to delete a template unless all the passes for that template are deleted first. That is what this utility does.

```
./delete-all-passes-for-templateId -templateId 12345 -authKey xxxxxxxxx
```

Instead of suppling the authKey at the command-line, I prefer to use a local proxy to add my Authorization Header onto the outgoing request. Refer to https://www.charlesproxy.com/documentation/tools/rewrite/

Here's how you can direct the program to send requests through a local proxy:
```
http_proxy=127.0.0.1:8888 ./delete-all-passes-for-templateId -templateId 12345
```


Also see:
```
./delete-all-passes-for-templateId --help
```