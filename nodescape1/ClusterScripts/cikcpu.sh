#!/bin/sh
/home/hankd/NODESCAPE/epacsedon cik cput `sensors | grep 'Core 0' |  awk '/^/{printf "%d", $3'}`
