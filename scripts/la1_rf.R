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

pdf(paste0(symbol, ".pdf"))

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
#head(df)
data_ex <- data.frame(Time=df$CloseTime, Close=df$High)
data_ex$Time <- as_datetime(data_ex$Time / 1000)

orig <- data_ex$Close
#orig <- orig[1:(dim(data_ex)[1]),2]
N <- length(orig)
ts <- orig
p <- 40
X <- matrix(rev(ts[1:p]))
for (i in 2:(N-p+1)) {
    X <- cbind(X, rev(ts[i:(i+p-1)]))
}

# norm(x, "F") x - matrix
norms <- c()
for (i in 1:(N-p)) {
    norms <- c(norms, norm(as.matrix(X[,i]-X[,N-p+1]), "F"))
}

nearest <-min(norms) 
eps <- 0
step <- nearest*0.01
num_neighbors <- 0
while (num_neighbors < 3 * (p+1)){
    ws <- which(norms < (nearest + eps))
    eps <- eps + step
    num_neighbors <- length(ws)
}

if (length(norms) == 0) {
    print("neighbors == 0")
}

Y <- X[1,(ws+1)]
A <- t(X[,ws])
df <- as.data.frame(A)
df <- cbind(Y, df)
library(randomForest)
set.seed(0)
rfManyReg <- randomForest(Y ~ ., data=df)
#print(rfManyReg)

x_t <- ts[(N-p+1):(N)]
pr <- data.frame(t(rev(x_t)))
colnames(pr) <- colnames(df)[2:dim(df)[2]]
y <- predict(rfManyReg, pr)

#print(data_ex[N+1,2])
#print(y)
tss <- ts
#N <- p
for (i in 1:p){
    x_t <- tss[(N-p+1):(N)]
    pr <- data.frame(t(rev(x_t)))
    colnames(pr) <- colnames(df)[2:dim(df)[2]]
    y <- predict(rfManyReg, pr)
    tss <- c(tss, y)
    N <- N+1
}
#png(file=paste0(symbol, ".png")
matplot(data.frame(tss, c(data_ex[,2], rep(NA, p))), type = "l", col = c('green', 'red'),
        ylab="price", xlab="time")
abline(v=N-p, col='red')
#print(paste("p =",p, "success"))
max_price <- max(ts[(length(ts)-p):length(ts)])
min_price <- min(ts[(length(ts)-p):length(ts)])

price <- ts[length(ts)-p]
eps <- price * 0.003
if (price < min_price + eps && price < max_price - eps){ # && price - eps < min_price) {
    cat(paste("1", price, min_price, max_price))
} else if (price > min_price + eps) {
    cat(paste("-1", price, min_price, max_price))
} else {
    cat(paste("0", price, min_price, max_price))
}
