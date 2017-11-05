# discordice

![Docker Hub build badge](https://dockeri.co/image/justinian/discordice)

discordice is a golang webservice to implement a Discord !roll command,
taking any dice argument supported by my [dice](https://github.com/justinian/dice)
library. 

## Running the service

The service is most easily run as a docker container. The only configuration necessary
is your incoming webhook integration URL.

```bash
docker run -d -e DISCORDICE_TOKEN="<your Discord token>" -p 8000:8000 --name=discordice justinian/discordice
```

See the docker container builds at https://registry.hub.docker.com/u/justinian/discordice

To add my instance of discordice to your server, [use this link][1].

[1]: https://discordapp.com/api/oauth2/authorize?client_id=232229987145482261&scope=bot&permissions=11264
