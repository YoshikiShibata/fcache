# frcache

`frcache` is a simple function-result cache: a supplied function will be executed periodically and its result will be cached for later queries.

When `Get` is called, the supplied function will be invoked to get the lastest result. But when the function will not return within a specified timeout, then any previously cached result will be returned.
