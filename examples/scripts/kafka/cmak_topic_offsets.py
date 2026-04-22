import json
from pathlib import Path

"""
Output is comma-separated latestOffset per partition (0..n-1), suitable for
`vkconfig microbatch --offset`. Refresh the JSON snapshot next to this file:

  curl -sS -u 'USER:PASS' \\
    'http://2kafka01.rtty.in:8888/api/status/2kafka/topicIdentities' \\
    -o examples/kafka/2kafka_topicIdentities.json
TO-DO: the curl step needs to be scripted later
"""

IDS_KEY = 'topicIdentities'
PART_IDS_KEY = 'partitionsIdentity'
PART_NUM_KEY = 'partNum'
PARTITIONS_KEY = 'partitions'
CMAK_TOPIC_IDS_RESPONSE = Path(__file__).resolve().parent / '2kafka_topicIdentities.json'


def get_cmak_topic_offsets(topic: str) -> str:
    with open(CMAK_TOPIC_IDS_RESPONSE, encoding='utf-8') as f:
        payload = json.load(f)

    identities = payload.get(IDS_KEY, [])
    for ident in identities:
        if ident.get('topic') != topic:
            continue

        partitions = ident.get(PART_IDS_KEY, [])
        expected = ident.get(PARTITIONS_KEY)
        if expected is not None and len(partitions) != int(expected):
            raise ValueError(
                f'{topic!r}: {PART_IDS_KEY} has {len(partitions)} entries, '
                f"topic metadata says partitions={expected}"
            )

        ordered = sorted(partitions, key=lambda p: p[PART_NUM_KEY])
        part_nums = [p[PART_NUM_KEY] for p in ordered]
        want = list(range(len(ordered)))
        if part_nums != want:
            raise ValueError(
                f'{topic!r}: expected partNum 0..{len(ordered) - 1}, got {part_nums!r}'
            )

        return ','.join(str(p['latestOffset']) for p in ordered)

    raise ValueError(f'topic not found')


if __name__ == '__main__':
    print(get_cmak_topic_offsets('click_avro'))
