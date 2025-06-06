package rate

import "time"

// nowfn is a package-level variable that provides the current time.
// Using a variable instead of directly calling time.Now allows for
// easier testing by making it possible to mock the time
// functionality. In production, this defaults to time.Now.
var nowfn = time.Now
