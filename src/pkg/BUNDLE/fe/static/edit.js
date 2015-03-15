function ajaxDone(data) {
        window.location = AdminURL;
}

function ajaxError(data) {
        window.location = AdminURL;
}

function onRevoke(event) {
        event.preventDefault();
        $.ajax({
                url: '/api/revoke?s=' + $('#f_slot').val(),
                success: ajaxDone,
                error: ajaxError,
                dataType: 'json',
        });
}

function onUpdate(event) {
        event.preventDefault();
        n = $.URLEncode($('#f_name').val());
        e = $.URLEncode($('#f_email').val());
        a = $.URLEncode($('#f_addr').val());
        $.ajax({
                url: '/api/update?s='+$('#f_slot').val()+"&n="+n+"&e="+e+"&a="+a,
                success: ajaxDone,
                error: ajaxError,
                dataType: 'json',
        });
}

$(document).ready(function(){
        mainReady();

        $('#f_id').attr("disabled", true);
        $('#f_id').addClass("disabled");

        $('#f_dk').attr("disabled", true);
        $('#f_dk').addClass("disabled");
        $('#f_ak').attr("disabled", true);
        $('#f_ak').addClass("disabled");

        $('#f_hk').attr("disabled", true);
        $('#f_hk').addClass("disabled");

        $('#f_sk').attr("disabled", true);
        $('#f_sk').addClass("disabled");

        $('#f_cancel').click(function(){ window.location = AdminURL; });
        $('#f_revoke').click(onRevoke);
        $('#f_update').click(onUpdate);
});
