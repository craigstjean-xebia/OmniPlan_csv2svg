#!/bin/bash

 ./omniplan_csv2svg -o examples/Sample.svg examples/Sample.csv
 ./omniplan_csv2svg -zoom 1 -o examples/SampleZ1.svg examples/Sample.csv
 ./omniplan_csv2svg -zoom 1 -level 5 -o examples/SampleZ1L5.svg examples/Sample.csv
 ./omniplan_csv2svg -level 5 -w 2000 -o examples/SampleL5W2000.svg examples/Sample.csv
 ./omniplan_csv2svg -t 2 -level 5 -o examples/SampleT2L5.svg examples/Sample.csv
 ./omniplan_csv2svg -t 4 -level 5 -o examples/SampleT4L5.svg examples/Sample.csv
 