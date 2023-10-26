document.addEventListener("DOMContentLoaded", function () {
    const lazyLoadImages = document.querySelectorAll("img.lazy");
    let lazyLoadThrottleTimeout;

    function lazyLoad() {
        if (lazyLoadThrottleTimeout) {
            clearTimeout(lazyLoadThrottleTimeout);
        }

        lazyLoadThrottleTimeout = setTimeout(function () {
            lazyLoadImages.forEach(function (img) {
                img.src = img.dataset.src;
                img.classList.remove('lazy');
            });
            if (lazyLoadImages.length === 0) {
                window.removeEventListener("load", lazyLoad)
            }
        }, 20);
    }

    window.addEventListener("load", lazyLoad);
});