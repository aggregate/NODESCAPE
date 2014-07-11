#!/bin/sh
/home/admin/NODESCAPE/epacsedon emcluster cput `sensors | grep 'temp1' |  awk '/^/{printf "%d", $2'}`
