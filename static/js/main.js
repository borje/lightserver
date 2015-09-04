$(document).ready(function() {
    $.getJSON("info", function(data) {

        $("#githash").append(data.version);
        $("#buildtime").append(data.buildTime);
    });

	$.getJSON("status", function(data) {
		var tableContent = ""
		$.each(data, function (i, obj) {
			var date = new Date(obj.Time);
			var dateString = "";
			if (!isToday(date)) {
				dateString = date.getFullYear() + "-" + date.getMonth() + "-" + date.getDate() + " ";
			}
			dateString += " " + date.getHours() + ":" + date.getMinutes();
			tableContent += "<tr><td>" + dateString + "</td><td>" + obj.Device+ "</td>";
			tableContent += "<td>" + (obj.Action == 1 ? "Off" : "On") + "</td></tr>";
		});
		$("#queue").append(tableContent);

	}); 
});

function isToday(date) {
	var today = new Date();
	return today.getDate() == date.getDate();
}
