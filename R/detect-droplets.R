library(tidyr)
library(dplyr)
library(Rcpp)
library(ggplot2)
sourceCpp("src/find-droplets.cpp")

raw <- read.csv("data/converted_from_qlb/009 plasmid_A01_RAW.qlb.v16le.csv")
raw$measurements <- 1:nrow(raw)
raw <- as_data_frame(raw)

amplitude <- read.csv("data/quantasoft_data/009 plasmid_2016-10-07-14-59/009 plasmid_A01_Amplitude.csv")

background <- mean(slice(raw, 1:5E5)$ch2)
sd <- sd(slice(raw, 1:5E5)$ch2)
sds <- 15

# this might be wrong !!!!
# DANGER !!!!
# !!!!!
calc_auc <- function(y) {
    id <- 1:length(y)
    x <- 1:length(y)
    sum(diff(x[id]) * zoo::rollmean(y[id],2))
}

result <- raw %>% 
  mutate(droplets = find_droplets_two_channels(ch1, ch2, 785, 600)) %>% 
  filter(droplets != 0) %>% 
  group_by(droplets) %>% 
  summarise(ch1w = length(ch1), ch2w = length(ch2), ch1h = max(ch1), ch2h = max(ch2))

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

singlets <- filter(result, (ch2w < 21 & ch2h < 1500) | (ch2w < 26 & ch2h > 1499))

# compare to reference
ggplot(amplitude, aes(x = Ch1.Amplitude)) + geom_density() + ggtitle("Reference")
ggplot(result, aes(x = ch1h)) + geom_density()  + ggtitle("Derived")

ggplot(amplitude, aes(x = Ch2.Amplitude)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Reference")
ggplot(result, aes(x = ch2h)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Derived")

ggplot(amplitude, aes(x = Ch2.Amplitude, y = Ch1.Amplitude)) + geom_point() + ggtitle("Reference")
ggplot(result, aes(x = ch2h, y = ch1h)) + geom_point() + ggtitle("Derived")

compensated <- singlets %>% 
    mutate(ch1hc = ch1h - ch2h * 0.375)

ggplot(compensated, aes(x = ch2h, y = ch1hc)) + geom_point() + ggtitle("Derived")