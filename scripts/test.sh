#!/usr/bin/env bash

while getopts "x:" flag; do 
  case $flag in 
    x) 
      echo $OPTARG
      ;;
  esac
done

echo SUPER_SECRET is $SUPER_SECRET
echo MORE is $MORE
echo MOAARRRR is $MOAARRRR

