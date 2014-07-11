#!/bin/sh
/home/admin/NODESCAPE/epacsedon axk cput `sensors | grep 'temp1' |  awk '/^/{printf "%d", $2'}`
