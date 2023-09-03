# merminder
Merminder (aka merge reminder) is a project that intend to keep track and notify about merge requests, so developers can remember of review them as soon as possible

## Usage
You can pull the Docker image and overide the `.merminder.yml`
```
docker run --name merminder -it -v <your config>:/app/.merminder.yml gabrielmusskopf/merminder:0.1
```

or, clone the repo, set your `.merminder.yml` and run
```shell
git clone https://github.com/gabrielmusskopf/merminder.git

//set your config

go run main.go
```

The git repo or docker image doesn't come with any defult configuration file, so it's mandatory you to set it up. Here's an example of `.merminder.yml`
```yaml
repository:
  host: "your gitlab host"
  token: "your gitlab token"

send:
  # on or off
  # the default is on
  notification: on
  templateFilePath: slack.tmpl
  webhookUrl: "your webhook url"

observe:
  # groups to observe. 
  # all porjects merge requests from this group will be monitored
  groups:
    - 1
  # projects to observe. 
  projects:
    - 2
  #every or at
  #every: 60s
  at: 
    - 09:35
    - 17:00
```

### TODO:
- This project still in process, so it still needs proper testing
- Maybe add a helm chart
- Some kind of review time history record and visualization
- Define another notification types with different messages and times
