#include <Rcpp.h>
using namespace Rcpp;


// [[Rcpp::export]]
NumericVector find_droplets(NumericVector ch1) {
    int magic_number = 785;
    int nrow = ch1.size();
    int magic_width = 7;
    NumericVector out(nrow);
    
    int droplet_number = 1;  
    
    for (int i = 0; i < nrow; i++) {
        if (ch1[i] > magic_number) {
            int j = i + 1;
            while (ch1[j] > magic_number & j < nrow) {
                j++;
            }
            if ((j - i) > magic_width) {
                for (int k = i; k < j; k++) {
                    out[k] = droplet_number;
                }
                droplet_number = droplet_number + 1;
            }
            i = j;
        }
    }
    
    return out;
}

// [[Rcpp::export]]
NumericVector find_droplets_two_channels(NumericVector ch1, NumericVector ch2, 
    int ch1threshold, int ch2threshold) {
    int nrow = ch1.size();
    int magic_width = 7;
    NumericVector out(nrow);
    
    int droplet_number = 1;  
    
    for (int i = 0; i < nrow; i++) {
        if (ch1[i] > ch1threshold & ch2[i] > ch2threshold) {
            int j = i + 1;
            while (ch1[j] > ch1threshold & ch2[j] > ch2threshold & j < nrow) {
                j++;
            }
            if ((j - i) > magic_width) {
                for (int k = i; k < j; k++) {
                    out[k] = droplet_number;
                }
                droplet_number = droplet_number + 1;
            }
            i = j;
        }
    }
    
    return out;
}