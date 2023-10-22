document.addEventListener("DOMContentLoaded", function() {
    const lazyLoadImages = document.querySelectorAll("img.lazy");
    let lazyLoadThrottleTimeout;

    function lazyLoad () {
        if(lazyLoadThrottleTimeout) {
            clearTimeout(lazyLoadThrottleTimeout);
        }

        lazyLoadThrottleTimeout = setTimeout(function() {
            // const scrollTop = window.pageYOffset;
            lazyLoadImages.forEach(function(img) {
                // if(img.offsetTop < (window.innerHeight + scrollTop)) {
                    img.src = img.dataset.src;
                    img.classList.remove('lazy');
                // }
            });
            if(lazyLoadImages.length === 0) {
                document.removeEventListener("scroll", lazyLoad);
                window.removeEventListener("resize", lazyLoad);
                window.removeEventListener("orientationChange", lazyLoad);
            }
        }, 20);
    }

    document.addEventListener("scroll", lazyLoad);
    window.addEventListener("resize", lazyLoad);
    window.addEventListener("orientationChange", lazyLoad);
});