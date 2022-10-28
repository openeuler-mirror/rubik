library(qcc)
library(readr)

spc_ipc_normal <- read.csv("Workspace/rubikAnalysis/tests/data/clickhouse/spc_ipc_abnormal.csv", sep="")
View(spc_ipc_normal)

measures <- spc_ipc_normal[,1]
# Specify the sample size
sample_size <- 3

# Get the total number of samples
number_of_samples <- length(measures)/sample_size

# Assign the measures to the corresponding sample
sample <- rep(1:number_of_samples, each = sample_size)

# Create a data frame with the measures and sample columns
df <- data.frame(measures, sample)

# Group the measures per sample
measure <- with(df, qcc.groups(measures, sample))

# View the grouped measures
head(measure)

# Specify the measures unit
measure_unit <- "lbs"  

# Create the x-bar chart
xbar <- qcc(measure, type = "xbar", data.name = measure_unit)

# Specify the warning limits (2 sigmas)
(warn.limits.2 = limits.xbar(xbar$center, xbar$std.dev, xbar$sizes, 2))

# Specify the warning limits (1 sigmas)
(warn.limits.1 = limits.xbar(xbar$center, xbar$std.dev, xbar$sizes, 1))

# Plot the x-bar chart
plot(xbar, restore.par = FALSE)

# Plot the warning limit lines
abline(h = warn.limits.2, lty = 2, col = "blue")
abline(h = warn.limits.1, lty = 2, col = "lightblue")

# Get the summary for x-bar chart
summary(xbar)

# Create the R-chart
r_chart <- qcc(measure, type = "R", data.name = measure_unit)

# Get the summaries for R-chart
summary(r_chart)


# Specify the lower control limit
LSL <- as.numeric(2.57704)
 
# Specify the upper control limit
USL <- as.numeric(2.7223)

# Specify the target
Target <- as.numeric(2.6497)

# Plot the process capability chart
process.capability(xbar, spec.limits = c(LSL, USL), target = Target)
