#! /usr/bin/sh
# DO NOT EDIT. Generated by terravalet.
# terravalet_output_format=2
#
# This script will move 10 items.

set -e

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.localhostnames_public["anka"]' \
    'aws_route53_record.anka'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.localhostnames_public["anka-api"]' \
    'aws_route53_record.anka-api'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.localhostnames_public["anka-test"]' \
    'aws_route53_record.anka-test'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.localhostnames_public["artifactory"]' \
    'aws_route53_record.artifactory'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.loopback["anka"]' \
    'aws_route53_record.anka_loopback'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.loopback["anka-api"]' \
    'aws_route53_record.anka_api_loopback'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.loopback["anka-test"]' \
    'aws_route53_record.anka_test_loopback'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.loopback["artifactory"]' \
    'aws_route53_record.artifactory_loopback'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.private["anka"]' \
    'aws_route53_record.anka_private'

terraform state mv -lock=false -state=local.tfstate \
    'aws_route53_record.private["artifactory"]' \
    'aws_route53_record.artifactory_private'
