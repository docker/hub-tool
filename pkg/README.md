# Using tool as lib

You can use this tool as a library to make actions to Docker Hub.

## Examples

### Login

```
username := toto
password := pass

hubClient, err := hub.NewClient(
	hub.WithHubAccount(username),
	hub.WithPassword(password))
if err != nil {
	log.Fatalf("Can't initiate hubClient | %s", err.Error())
}

//Login to retrieve new token to Hub
token, _, err := hubClient.Login(username, password, func() (string, error) {
	return "2FA required, please provide the 6 digit code: ", nil
})
if err != nil {
	log.Fatalf("Can't get token from Docker Hub | %s", err.Error())
}
```

After a successful login, it is quite easy to do any action possible and listed inside `pkg/` directory.

### Removing a tag

```
err = hubClient.RemoveTag("toto/myrepo", "v1.0.0")
if err != nil {
	log.Fatalf("Can't remove tag | %s", err.Error())
}
```
