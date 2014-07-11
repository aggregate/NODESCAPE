#!/bin/sh
/home/admin/NODESCAPE/epacsedon nak gput `~/bin/nvidia-smi -a -q |awk '/Temperature/ { printf "%d\n", $3 }'`

