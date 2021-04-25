library('httr')
library('jsonlite')
library('lubridate')

args = commandArgs(trailingOnly=TRUE)
#apikey <- 'PwR4lONtheOWGpQK0dBdH4yPx6AsGSl57xS8j396KeZAxcKEaaYbjaVh8s1F7B4b'
endpoint <- "https://api.binance.com"
url <- "/api/v3/klines"
interval <- args[1]
#startTime <- trunc(unclass(Sys.time() - 150 * 24 * 3600)) * 1000
#endTime <- trunc(unclass(Sys.time())) * 1000
symbol <- args[2]
limit <- 1000
URL <- paste0(endpoint, url, 
              '?symbol=', symbol, 
              '&interval=', interval,
              #'&startTime=', startTime,
              #             '&endTime=', endTime,
              '&limit=', limit)
r <- GET(URL)
body <- content(r, as="parsed", type="application/json") # decode json
v <- unlist(body)
df <- as.data.frame(t(matrix(as.double(v), nrow = 12)))
colnames(df) <- c('OpenTime', 'Open', 'High', 'Low', 'Close', 
                  'Volume', 'CloseTime', 'QuoteAssetVolume', 
                  'NumberOfTrades', 'TakerBuyBaseAV', 'TakerBuyQuoteAV',
                  'Ignore')
head(df)
data_ex <- data.frame(Time=df$OpenTime, Close=df$High)
data_ex$Time <- as_datetime(data_ex$Time / 1000)

orig <- data_ex
orig <- orig[1:(dim(data_ex)[1]),2]
N <- length(orig)
ts <- orig
p <- 120
X <- matrix(ts[1:p])
for (i in 2:(N-p+1)) {
    X <- cbind(X, ts[i:(i+p-1)])
}
# norm(x, "F") x - matrix
norms <- c()
for (i in 1:(dim(X)[2]-1)) {
    if (i != N-p+1) {
        norms <- c(norms, norm(as.matrix(X[,i]-X[,N-p+1]), "F"))
    }
}
nearest <-min(norms) 
eps <- 0
step <- nearest*0.25
num_neighbors <- 0
while (num_neighbors < 3 * (p+1)){
    ws <- which(norms < (nearest + eps))
    eps <- eps + step
    num_neighbors <- length(ws)
}

if (length(norms) == 0) {
    print("neighbors == 0")
}

Y <- X[dim(X)[1],(ws+1)]
A <- t(X[,ws])
m <- lm(Y ~ A)
summary(m)

predict <- function(x_t, coef) {
    y <- coef[1] + t(x_t) %*% coef[2:length(coef)]
    return(y)
}
x_t <- ts[(N-p+1):(N)]
a <- as.vector(m$coefficients)
y <- predict(x_t, a)

#print(data_ex[N+1,2])
print(y)
for (i in 1:p){
    y <- predict(ts[(N-p+1):(N)], a)
    ts <- c(ts, y)
    N <- N+1
}
matplot(data.frame(ts, c(data_ex[,2], rep(NA, p))), type = "l", col = c('green', 'red'),
        ylab="price", xlab="time")
abline(v=N-p, col='red')
print(paste("p =",p, "success"))
max(ts[(length(ts)-p):length(ts)])
min(ts[(length(ts)-p):length(ts)])

#high <- ts
#low <- ts
#matplot(data.frame(low, high), type = "l", col = c('red', 'green'),
#ylab="price", xlab="time")

