var API_URL = "/api/v1/";
$(function(){
	$("#submit_login").on("click", function(e) {
		var arr = $("#login_form").serializeArray(),
			data = {};
		for (var i = 0; i < arr.length; i++) {
			data[arr[i].name] = arr[i].value;
		}

        $.ajax({
            type: "POST",
            url: API_URL+"login",
            data: JSON.stringify(data),
            contentType: "application/json",
            processData: false
        }).done(function(d) {
			$("#login-modal").hide();
            $("#workplace").show();
            Cookies.set('token', d.token);
            getHistory();
            setInterval(getHistory, 5000);    
        }).fail(function (x) {
        });

		return false;
    });

    $("#send_link").on("click", function(e){
        var link = $("#link_value").val();
        $.ajax({
            type: "GET",
            url: API_URL+"audio",
            data: "link="+link,
            processData: false,
            headers: {"Authorization": " BEARER " + Cookies.get('token')}
        }).done(function(d) {
            $('<div id="job_alert" class="alert alert-success" role="alert">'+ d.jobID +'</div>').insertBefore("#link_container");
            $("#job_alert").delay(3000).slideUp('slow', function(){ $(this).remove(); } )
        }).fail(function (x) {
        });
    });
});



function getHistory() {
    $.ajax({
        type: "GET",
        url: API_URL+"history",
        processData: false,
        headers: {"Authorization": " BEARER " + Cookies.get('token')}
    }).done(function(d) {
        var html = "";
        for (var i = 0; i < d.history.length; i++) {
            var item = d.history[i];
            html += '<tr><th scope="row">' + (i + 1) +'</th>'
            html += '<td>'+ item.time +'</td>'
            html += '<td>'+ item.title +'</td>'
            html += '<td>'+ item.audio_link +'</td>'
            html += '<td>'+ item.status +'</td>'
            html += '</tr>'
        }
        $("#history_body").html(html);
        $('#history_container').show();
    }).fail(function (x) {
    });
}