# space-watcher

space-watcher is a Twitter bot that tweets automatically when Followings start Twitter Spaces.

## Build

```shell
git clone https://github.com/qitoi/space-watcher.git
cd space-watcher
go build github.com/qitoi/space-watcher/cmd/space-watcher
```

## Usage
### Setup

```shell
cp config.example.yaml config.yaml

sed -i -e 's/YOUR_CONSUMER_KEY/<YOUR_CONSUMER_KEY>/' config.yaml
sed -i -e 's/YOUR_CONSUMER_SECRET/<YOUR_CONSUMER_SECRET>/' config.yaml

./space-watcher --init
```

### Start

```shell
./space-watcher
```

## Limitation

- up to 100 Followings

## License

Apache License 2.0

```
Copyright 2021 qitoi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
