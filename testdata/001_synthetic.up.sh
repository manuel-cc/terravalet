#! /usr/bin/sh
# DO NOT EDIT. Generated by terravalet.
# terravalet_output_format=2
#
# This script will move 3 items.

set -e

terraform state mv -lock=false -state=local.tfstate \
    'module.ci.aws_batch_compute_environment.concourse_gpu_batch' \
    'aws_batch_compute_environment.concourse_gpu_batch'

terraform state mv -lock=false -state=local.tfstate \
    'module.ci.aws_instance.bar' \
    'aws_instance.bar'

terraform state mv -lock=false -state=local.tfstate \
    'module.ci.aws_instance.foo["cloud"]' \
    'aws_instance.foo["cloud"]'

