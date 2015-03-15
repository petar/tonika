function addResultOK(data) {
        $('#wait').hide();

        $('#f_msg').val(data.Msg);
        $('#f_done').click(function(event){ 
                event.preventDefault();
                window.location = AdminURL; 
        });

        $('#invite').show();
        $.scrollTo('#great',{duration:"400"});
}

function addResultError(data) {
        $('#wait').hide();
        $('#error').show();
}

function onAlright(event) {
        event.preventDefault();
        $('#f_name').attr("disabled", true);
        $('#f_email').attr("disabled", true);
        $('#f_ok').attr("disabled", true);
        $('#f_ok').addClass("disabled");
        $('#f_ok').blur();
        $('#f_cancel').attr("disabled", true);
        $('#f_cancel').hide();

        n = $.URLEncode($('#f_name').val());
        e = $.URLEncode($('#f_email').val());
        $('#wait').show();
        $.ajax({
                url: '/api/add?n='+n+'&e='+e,
                success: addResultOK,
                error: addResultError,
                dataType: 'json',
        });
}

$(document).ready(function(){
        mainReady();
        $('#f_name').removeAttr("disabled");
        $('#f_email').removeAttr("disabled");
        $('#f_cancel').removeAttr("disabled");
        $('#f_ok').removeAttr("disabled");
        $('#f_ok').click(onAlright);
        $('#f_cancel').click(function(){ window.location = AdminURL; });
});
