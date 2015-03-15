function ajaxDone(data) {
        $('#f_text').text(data)
}

function ajaxError() {
}

function refresh() {
        $.ajax({
                url: '/api/monitor',
                success: ajaxDone,
                error: ajaxError,
                dataType: 'text',
        });
}

$(document).ready(function(){
        mainReady();
        window.setInterval('refresh()', 1000)
});
