function isScrolledIntoView(el) {
	var rect = el.getBoundingClientRect();
	var elemTop = rect.top;
	var elemBottom = rect.bottom;
	var isVisible = (elemTop >= 0) && (elemTop <= window.innerHeight);
	return isVisible;
}

function ready(fn) {
	if (document.attachEvent ? document.readyState === "complete" : document.readyState !== "loading"){
		fn();
	} else {
		document.addEventListener('DOMContentLoaded', fn);
	}
}

function lazyload() {
	var images = document.querySelectorAll(".lazyload");
	for (var i = 0; i < images.length; i++) {
		if (isScrolledIntoView(images[i])) {
			var src = images[i].getAttribute("data-src");
			if (images[i].tagName.toLowerCase() == "video") {
				images[i].children[0].setAttribute("src", src);
				images[i].load();
			} else {
				images[i].style.backgroundImage = 'url("'+src+'")';
			}
			images[i].classList.remove("lazyload");
			images[i].removeAttribute("data-src");
		}
	}
}

document.addEventListener("scroll", lazyload);
document.addEventListener("turbolinks:render", lazyload);
document.addEventListener("turbolinks:load", lazyload);
document.addEventListener("load", lazyload);

ready(lazyload);
