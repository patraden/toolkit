# Setup env

Export once; then run any command below. Change `TOPIC` to switch topics.

```zsh
export KAFKA_REGISTER="http://registry-2kafka.rtty.in:8081"
export TOPIC="onclick_conversion_context"
export BOOTSTRAP_HOST="2kafka01.rtty.in:9092"
```

# Read the latest message from a topic

Consumer mode (`-C`), one message from end of topic (`-o end -c 1`), Avro value with schema registry, output piped to `jq`.

```zsh
kcat -b "$BOOTSTRAP_HOST" -C -t "$TOPIC" -o end -c 1 -s avro -r "$KAFKA_REGISTER" | jq
```

# Read last 1000 messages and filter by attribute

Consume up to 1000 messages from offset `-10000` (`-o -10000 -c 1000`), Avro value, then filter with `jq`. Add `-C` for consumer mode if needed.

```zsh
kcat -b "$BOOTSTRAP_HOST" -C -t "$TOPIC" \
     -r "$KAFKA_REGISTER" \
     -s value=avro \
     -o -10000 \
     -c 1000 \
  | jq -c 'select(.participants != null)'
```