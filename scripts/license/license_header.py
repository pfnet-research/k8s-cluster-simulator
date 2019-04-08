# Copyright 2019 Preferred Networks, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

LICENSE_HEADER = [
    'Copyright 2019 Preferred Networks, Inc.',
    '',
    'Licensed under the Apache License, Version 2.0 (the "License");',
    'you may not use this file except in compliance with the License.',
    'You may obtain a copy of the License at',
    '',
    '    http://www.apache.org/licenses/LICENSE-2.0',
    '',
    'Unless required by applicable law or agreed to in writing, software',
    'distributed under the License is distributed on an "AS IS" BASIS,',
    'WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.',
    'See the License for the specific language governing permissions and',
    'limitations under the License.',
]

LICENSE_HEADER_MODIFICATION = [
    'Modifications copyright 2019 Preferred Networks, Inc.',
] + LICENSE_HEADER[1:]


def license_header(comment, modification=False):
    """
    Returns a license header.
    Each line is preceded by `comment` string, and a single space if the line is not empty.
    """

    license = LICENSE_HEADER_MODIFICATION if modification else LICENSE_HEADER
    return '\n'.join(
        "{} {}".format(comment, l) if len(l) > 0 else comment
        for l in license
    )


def has_license_header(path, license_header):
    """
    Returns whether the file at `path` contains `license_header` string.
    """

    return license_header in path.open().read()
