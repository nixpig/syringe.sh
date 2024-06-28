#!/usr/bin/env bash

echo args: 
while getopts "x:" flag; do 
  case $flag in 
    x) 
      echo $OPTARG
      ;;
  esac
done

echo HANGER is $HANGER
echo REBELBASE is $REBELBASE

