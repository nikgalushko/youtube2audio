$(function(){
    var API_URL = "/api/v1/";
	$("#login_form").on("submit", function(e) {
		var arr = $(this).serializeArray(),
			data = {};
		for (var i = 0; i < arr.length; i++) {
			data[arr[i].name] = arr[i].value;
		}
        console.log(JSON.stringify(data));
        $.ajax({
            type: "POST",
            url: API_URL+"login",
            data: JSON.stringify(data),
            contentType: "application/json",
            processData: false
        }).done(function(d) {
			document.cookie = "login" + "=" + escape(data['login']);
			$("#login_container").hide();
			$("#workplace").show();
			loadRecords("emotions="+currentEmotion);
        }).fail(function (x) {
            alert(x);
        });

		return false;
	})
});