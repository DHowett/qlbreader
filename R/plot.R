# set the working directory to the qlbreader directory if it's not already
# setwd("~/path/to/qlbreader")

# install packages if not already installed
# install.packages(c("ggplot2", "tidyr", "dplyr", "stringr"))

library(ggplot2)
library(tidyr)
library(dplyr)
library(stringr)

# Look at complete files converted from qlb ------------------------------------
files <- list.files("data/converted_from_qlb")
files <- files[str_detect(files, "csv")]

read_full_file <- function(f) {
  x <- read.csv(paste0("data/converted_from_qlb/", f))
  file_name <- str_replace(f, "\\.csv", "")
  x$file <- file_name
  x$measurements <- 1:nrow(x)
  x <- gather(x, key = "channel", value = "fluorescence", ch1, ch2)
  x
}

x <- bind_rows(lapply(files, read_full_file))

p <- function(d) {
  ggplot(d, aes(x = measurements, y = fluorescence)) +
    geom_line() +
    facet_grid(file ~ channel,  scales = "free_y")
}

# plot subset of data
sample_n(x, nrow(x) / 40) %>% 
  p()

# plot zoomed in 
slice(x, 2.001E6:2.004E6) %>% 
  p()

# plot background subtracted
ch1 <- filter(x, channel == "ch1")
ch1bg <- mean(ch1$fluorescence[1:100000])
ch1$fluorescence <- ch1$fluorescence - ch1bg

sample_n(ch1, nrow(ch1) / 40) %>% 
  p()