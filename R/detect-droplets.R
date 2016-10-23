library(tidyr)
library(dplyr)
library(Rcpp)
library(ggplot2)
sourceCpp("src/find-droplets.cpp")

raw <- read.csv("data/converted_from_qlb/009 plasmid_A01_RAW.qlb.v16le.csv")
raw$measurements <- 1:nrow(raw)
raw <- as_data_frame(raw)

amplitude <- read.csv("data/quantasoft_data/009 plasmid_2016-10-07-14-59/009 plasmid_A01_Amplitude.csv")
amplitude <- as_data_frame(amplitude)

background <- mean(slice(raw, 1:5E5)$ch2)
sd <- sd(slice(raw, 1:5E5)$ch2)
sds <- 15

# this might be wrong !!!!
# DANGER !!!!
# !!!!!
calc_auc <- function(y, channel_mean) {
    # rise = max(y1, y2) - min(y1, y2)
    # rise_area = (rise * 1) * 1/2
    # stalk = min(y1, y2)
    # stalk_area = stalk * 1
    # area = rise_area + stalk_area
    y <- y - channel_mean
    sum(y) - ((y[1] + y[length(y)]) / 2)
}

# look at just droplets
just_droplets <- raw %>% 
    mutate(droplets = find_droplets_two_channels(ch1, ch2, 750, 600)) %>% 
    filter(droplets != 0)  

just_droplets %>% 
    mutate(measurements = 1:nrow(just_droplets)) %>% 
    ggplot(aes(x = measurements, y = ch2)) + geom_line() + xlim(c(0, 1000))

# filter out droplets
result <- raw %>% 
  mutate(droplets = find_droplets_two_channels(ch1, ch2, 750, 600)) %>% 
  filter(droplets != 0) %>% 
  group_by(droplets) %>% 
  summarise(ch1w = length(ch1), ch2w = length(ch2), 
      ch1h = max(ch1), ch2h = max(ch2), 
      ch1a = calc_auc(ch1, 598), ch2a = calc_auc(ch2, 422))

ggplot(result, aes(x = ch1h, y = ch2h)) + geom_point()
ggplot(result, aes(x = ch1w, y = ch1h)) + geom_point()
ggplot(result, aes(x = ch2w, y = ch2h)) + geom_point() + xlim(c(0, 40))

# trim extra wide ones to get singlets?
ggplot(result, aes(x = ch1w)) + geom_bar()
ggplot(result, aes(x = ch2w)) + geom_bar()
ggplot(result, aes(x = ch1w)) + geom_density()
ggplot(result, aes(x = ch2w)) + geom_density()
ggplot(result, aes(x = ch1w, y = ch1h)) + geom_point() +
    xlim(c(0, 40))

ggplot(result, aes(x = ch2w, y = ch2h)) + geom_point() +
    xlim(c(0, 40))

# look at raw doublet 
# filter(result, ch1w == 30, ch1h > 1400, ch1h < 1500)
# 
# raw %>% 
#     mutate(droplets = find_droplets_two_channels(ch1, ch2, 750, 600)) %>% 
#     filter(droplets == 14182)  %>% 
#     ggplot(aes(x = measurements, y = ch1)) + geom_point()

# get singlets
singlets <- filter(result, (ch2w < 21 & ch2h < 1750) | (ch2w < 27 & ch2h > 1750))

# compare to reference
ggplot(amplitude, aes(x = Ch1.Amplitude)) + geom_density() + ggtitle("Reference")
ggplot(result, aes(x = ch1h)) + geom_density()  + ggtitle("Derived")

ggplot(amplitude, aes(x = Ch2.Amplitude)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Reference")
ggplot(result, aes(x = ch2h)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Derived")

ggplot(amplitude, aes(x = Ch2.Amplitude, y = Ch1.Amplitude)) + geom_point() + ggtitle("Reference")
ggplot(result, aes(x = ch2h, y = ch1h)) + geom_point() + ggtitle("Derived")

compensated <- singlets %>% 
    mutate(
        ch1h = ch1h - 597.6,
        ch2h = ch2h - 421.63,
        ch1hc = 1.03 * ch1h + ch2h * -0.39, 
        ch2hc = ch1h * -0.36 + 1.96 * ch2h)

ggplot(compensated, aes(x = ch2hc, y = ch1hc)) + geom_point() + ggtitle("Derived")