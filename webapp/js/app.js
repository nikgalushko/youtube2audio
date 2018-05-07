$(function(){
    var API_URL = "/api/v1/";
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
        }).fail(function (x) {
            alert(x);
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
            $("#job_alert").delay(5000).slideUp('slow', function(){ $(this).remove(); } )
        }).fail(function (x) {
            //alert(x);
        });
    });
});