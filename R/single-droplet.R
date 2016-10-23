# droplet # 7441

library(tidyr)
library(dplyr)
library(Rcpp)
library(ggplot2)
sourceCpp("src/find-droplets.cpp")

raw <- read.csv("data/converted_from_qlb/009 plasmid_A01_RAW.qlb.v16le.csv")
raw$measurements <- 1:nrow(raw)
raw <- as_data_frame(raw)

result <- raw %>% 
    mutate(droplets = find_droplets_two_channels(ch1, ch2, 750, 600)) %>% 
    filter(droplets == 7441)

result2 <- raw %>% 
    mutate(droplets = find_droplets_two_channels(ch1, ch2, 750, 600)) %>% 
    filter(measurements > 2299130, measurements < 2299175)

ggplot(result, aes(x = measurements, y = ch1)) + geom_point()
ggplot(result, aes(x = measurements, y = ch2)) + geom_point()

ggplot(result2, aes(x = measurements, y = ch1 - 598)) + geom_point()
ggplot(result2, aes(x = measurements, y = ch2 - 422)) + geom_point()


amplitude <- read.csv("data/quantasoft_data/009 plasmid_2016-10-07-14-59/009 plasmid_A01_Amplitude.csv")
amplitude <- as_data_frame(amplitude)
filter(amplitude, Ch1.Amplitude > 700)

max(result$ch1)
max(result$ch2)

max(result$ch1) / 756.2719
max(result$ch2) / 2712.859

ch1 <- max(result$ch1)
ch2 <- max(result$ch2)

ch1 * 1.034385 + ch2 * -0.3602588
ch1 * -0.394962 + ch2 * 1.962357
