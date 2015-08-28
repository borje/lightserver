
$(document).ready(function() {
    $.getJSON("info", function(data) {

        $("#githash").append(data.version);
        $("#buildtime").append(data.buildTime);
    });

});

