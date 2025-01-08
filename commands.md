ssh command

```
ssh -i "key.pem" ec2-user@ec2-18-144-90-118.us-west-1.compute.amazonaws.com
```

Build command:

```
GOOS=linux GOARCH=amd64 go build -o main .
```

Transfer filers:


```
cd ~/dev 
scp -i key.pem ~/GolandProjects/wineterfest/main ec2-user@ec2-18-144-90-118.us-west-1.compute.amazonaws.com:/home/ec2-user
scp -i key.pem -r ~/GolandProjects/wineterfest/html ~/GolandProjects/wineterfest/main ec2-user@ec2-18-144-90-118.us-west-1.compute.amazonaws.com:/home/ec2-user
```

Run New Binary

first kill binary and delete all the files.

```
sudo lsof -i :8080
sudo kill -9 <PID>
```

run it:

```
chmod +x ./main
sudo systemctl daemon-reload
sudo systemctl restart goapp
```

logs:

```
sudo journalctl -u goapp -f
```

URL: http://18.144.90.118:8080/signup
