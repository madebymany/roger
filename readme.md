###Roger

Roger is a job scheduler, similar to cron but will run in the foreground. It expects to be run as the correct user, we tend to use it paired with runit so we can execute it with chpst

```
e.g. chpst -u user:group roger {cmd}
```
