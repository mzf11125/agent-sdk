import re

from agent_sdk.version import VERSION


def test_version_is_semver_like():
    assert re.match(r"^\d+\.\d+\.\d+$", VERSION)
