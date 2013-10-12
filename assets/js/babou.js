/**
	babou.js
	
	Provides increased interactivity for the `babou` frontend.
	This code is distributed under the BSD license which governs
	the entire `babou` project.

	You may read more about the license at `http://github.com/drbawb/babou`

	(c) Robert Straw 2013, All Rights Reserved.
*/

$(document).ready(function(){
	// If episode-name is clicked load latest episodes.
	var templateCache = {};
	templateCache["search_episodes"] = Handlebars.compile($("#t-search-episodes").html());
	templateCache["search_series"] = Handlebars.compile($("#t-search-series").html());

	$('#search_episodes').change(function(evt) {
		$.ajax({
			headers: { 
				Accept : "application/json; charset=utf-8",
				"Content-Type": "text/plain; charset=utf-8"
			},
			url: "/torrents/tv/episodes"
		}).done(function(data) {
			console.log("Preparing episodes ...")
			// wrap data in object
			var parsedData = JSON.parse(data);
			var context = {"episodes": parsedData};
			
			// compile template
			var renderedHTML = templateCache["search_episodes"](context);

			$('#torrent-list').fadeOut(200,function(){ 
				$(this).html(renderedHTML).fadeIn(200);
			});
		});
	});

	// If series-name is clicked load latest series
	$('#search_series').change(function(evt) {
		$.ajax({
			headers: { 
				Accept : "application/json; charset=utf-8",
				"Content-Type": "text/plain; charset=utf-8"
			},
			url: "/torrents/tv/series"
		}).done(function(data) {
			console.log("Preparing ...")
			// wrap data in object
			var parsedData = JSON.parse(data);
			var context = {"series": parsedData};
			for (var i = 0; i < context.series.length; i++) {
				// split episodes into head & tail for each series.
				context.series[i].head = _.head(context.series[i].episodes);
				context.series[i].tail = _.tail(context.series[i].episodes);
				context.series[i].numEpisodes = function() {
					if ("episodes" in context.series[i]) {
						return context.series[i].episodes.length;	
					} else {
						return 0
					}
				}();
			}

			// compile template
			var renderedHTML = templateCache["search_series"](context);
			$('#torrent-list').fadeOut(200,function(){
				$(this).html(renderedHTML).fadeIn(200);
			});
		});
	});
});