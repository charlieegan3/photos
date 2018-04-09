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

(function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
(i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
})(window,document,'script','https://www.google-analytics.com/analytics.js','ga');
ga('create', 'UA-46126659-2', 'auto');

var tracking = {
  log: function(thing, action, label) {
	ga("send", "event", thing, action, label, null);
  },

  logOutboundLinkClick: function(url) {
	console.log(url);
	if (ga.loaded) {
	  tracking.log("outbound", "click", url);
	}
	var win = window.open(url, '_blank');
	win.focus();
  },

  attachToLinks: function() {
	var array = [];
	var links = document.getElementsByTagName("a");
	for(var i=0; i<links.length; i++) {
	  if (links[i].getAttribute("href").includes("http")) {
		links[i].setAttribute("onclick", "tracking.logOutboundLinkClick('" + links[i].href + "'); return false;");
	  }
	}
  }
}

document.addEventListener("turbolinks:load", function(event) {
  ga("set", "location", location.pathname);
  ga("send", "pageview");
  tracking.attachToLinks();
});

document.addEventListener("scroll", lazyload);
document.addEventListener("turbolinks:render", lazyload);
document.addEventListener("turbolinks:load", lazyload);
document.addEventListener("load", lazyload);

ready(lazyload);
