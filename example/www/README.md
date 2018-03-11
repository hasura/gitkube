# Quickstart - Build your own Docker image


```bash
$ docker build -t nodejs-express:<tag> .
```

Run the Docker container using the command below.

```bash
$ docker run -d -p 8080:8080 nodejs-express:<tag>
```

# Quickstart - git based pipeline

Follow the steps mentioned below for git based pipeline

1. Ensure that you have a git project
2. Edit `app/src/server.js`
3. Commit your changes

    ```bash
    $ git add .
    $ git commit -m "message"
    ```

4. Push the changes to git

    ```bash
    $ git push <remote> master
    ```

# Advanced usage

### **Port**

Default Port for application is `8080` .
