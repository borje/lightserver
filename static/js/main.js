$(document).ready(function() {
    $.getJSON("info", function(data) {

        $("#githash").append(data.version);
        $("#buildtime").append(data.buildTime);
    });

    $.getJSON("status", function(data) {
        var tableContent = ""
            $.each(data, function (i, obj) {
                var date = new Date(obj.Time);
                var m = moment(date);
                tableContent += "<tr><td>" + m.format("YYYY-MM-DD HH:mm:ss") + "</td><td>" + obj.Device+ "</td>";

                tableContent += "<td>" + (obj.Action == 1 ? "Off" : "On") + "</td></tr>";
            });
        $("#queue").append(tableContent);

    }); 
    $('#calendar').fullCalendar({
        header: {
            left: 'prev,next today',
            center: 'title',
            right: 'month,agendaWeek,agendaDay',
        },
        allDaySlot: false,
        //defaultDate: '2015-10-01',
        defaultView: 'agendaWeek',
        firstDay: 1,
        //height: 550,
        lang: 'es',
        //titleFormat: 'YYYY-MM-DD',
        //slotDuration: '02:00:00',
        timeFormat: 'H(:mm)',
        events: 'schedule'
    });
});

function isToday(date) {
    var today = new Date();
    return today.getDate() == date.getDate();
}
