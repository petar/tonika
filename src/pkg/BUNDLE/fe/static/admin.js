function ajaxDone(data) {
        window.location = AdminURL;
}

function ajaxError(data) {
        window.location = AdminURL;
}

function onUpdate(event) {
        event.preventDefault();
        n = $.URLEncode($('#my_name').val());
        e = $.URLEncode($('#my_email').val());
        a = $.URLEncode($('#my_addr').val());
        $.ajax({
                url: '/api/myinfo?n=' + n + '&e=' + e + '&a=' + a,
                success: ajaxDone,
                error: ajaxError,
                dataType: 'json',
        });
}

$(document).ready(function(){
        mainReady();
        $('#f_add').click(function(){ window.location = AdminURL+"/add"; });
        $('#f_update').click(onUpdate);
});
