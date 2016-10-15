library(tidyr)
library(dplyr)
library(Rcpp)

raw <- read.csv("data/converted_from_qlb/009 plasmid_A01_RAW.qlb.v16le.csv")
raw$measurements <- 1:nrow(raw)
raw <- as_data_frame(raw)

amplitude <- read.csv("data/quantasoft_data/009 plasmid_2016-10-07-14-59/009 plasmid_A01_Amplitude.csv")

sample_n(raw, nrow(raw) / 400) %>% 
  ggplot(aes(x = measurements, y = ch2)) +
  geom_line() +
  geom_point(aes(color = droplets2 != 0)) +
    geom_hline(yintercept = background + 15*sd) 

slice(raw, 3.55E6:3.555E6) %>% 
  ggplot(aes(x = measurements, y = ch1)) +
  geom_line() +
  geom_point(aes(color = droplets2 != 0)) +
  geom_hline(yintercept = background + 15*sd)

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
ggplot(result, aes(x = ch2w, y = ch2h)) + geom_point()

# compare to reference
ggplot(amplitude, aes(x = Ch1.Amplitude)) + geom_density() + ggtitle("Reference")
ggplot(result, aes(x = ch1h)) + geom_density()  + ggtitle("Derived")

ggplot(amplitude, aes(x = Ch2.Amplitude)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Reference")
ggplot(result, aes(x = ch2h)) + geom_density() + xlim(c(0, 1E4)) + ggtitle("Derived")

# trim extra wide ones to get singlets?
ggplot(result, aes(x = ch1w)) + geom_bar()
ggplot(result, aes(x = ch2w)) + geom_bar()
ggplot(result, aes(x = ch1w)) + geom_density()
ggplot(result, aes(x = ch2w)) + geom_density()
ggplot(result, aes(x = ch1w, y = ch1h)) + geom_hex(bins = 100) +
  xlim(c(0, 40))

ggplot(result, aes(x = ch2w, y = ch2h)) + geom_hex(bins = 100) +
    xlim(c(0, 40))
