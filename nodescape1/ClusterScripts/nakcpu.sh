#!/bin/sh
/home/admin/NODESCAPE/epacsedon nak cput `cat /proc/acpi/thermal_zone/THRM/temperature | awk '/^/{printf "%d", $2'}`
