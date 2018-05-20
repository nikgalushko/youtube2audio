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
            setInterval(getHistory, 60000);    
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

    $("#rss_link").on("click", function(e){
        getRssLink();
        return false;
    })
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
            html += '<tr id="'+item.id+'"><th scope="row">' + (i + 1) +'</th>'
            html += '<td>'+ item.time +'</td>'
            html += '<td>'+ item.title +'</td>'
            html += '<td><a href='+item.audio_link+'>ссылка</a></td>'
            html += '<td>'+ item.status +'</td>'
            html += '<td><button onClick="deleteHistoryItem("'+item.id+'");" class="btn btn-outline-secondary" type="button">Уалить</button></td>'
            html += '</tr>'
        }
        $("#history_body").html(html);
        $('#history_container').show();
    }).fail(function (x) {
    });
}

function deleteHistoryItem(id) {
    $.ajax({
        type: "DELETE",
        url: API_URL+"delete_from_history/" + id,
        processData: false,
        headers: {"Authorization": " BEARER " + Cookies.get('token')}
    }).done(function(d) {
        $("#" + id).remove();
    }).fail(function (x) {
    });
}

function getRssLink() {
    $.ajax({
        type: "GET",
        url: API_URL+"generate_rss_link",
        processData: false,
        headers: {"Authorization": " BEARER " + Cookies.get('token')}
    }).done(function(d) {
        var uri = window.location.origin + API_URL + d.rss_link;
        window.open(uri, '_blank');
    }).fail(function (x) {
    });
}