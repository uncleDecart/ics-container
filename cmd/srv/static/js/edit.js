async function renderJSON(documentName) {
    var jsonTextarea = document.getElementById(documentName).value;
    var checkboxes = document.querySelectorAll('input[type="checkbox"]:checked');
    var name = document.getElementById('name').value;
    var values = Array.from(checkboxes).map(function (checkbox) {
        return checkbox.value;
    });

    var reqBody = {
        'Message': jsonTextarea,
        'config': {
            'Name': name,
            'PatchEnvelopes': values,
            'DriverType': 'dummy',
            'BackupType': 'nobackup',
        },
    }

    try {
        // Send the JSON string to the server
        const response = await fetch('/api/render', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(reqBody), 
        });

        // Parse the response JSON
        const result = await response.text();
        alert(result)

    } catch (error) {
        alert("Error sending or receiving data: " + error.message);
    }
}

function getJsonMap(containerSelector) {
    var inputPairs = document.querySelectorAll(containerSelector + ' p');
    var jsonMap = {};

    inputPairs.forEach(function (pair) {
        var key = pair.querySelector('input:nth-child(1)').value;
        var value = pair.querySelector('input:nth-child(2)').value;
        jsonMap[key] = value;
    });

    return jsonMap;
}

async function submit() {
    const name = document.getElementById('name').value;
    
    const checkboxContainer = document.getElementById('envelopesContainer');
    const checkboxes = checkboxContainer.querySelectorAll('input[type="checkbox"]:checked');
    const patchEnvelopes = Array.from(checkboxes).map(function (checkbox) {
        return checkbox.value;
    });

    const endpoint = document.getElementById('endpoint').value;

    const headers = document.getElementById('headersEditor').value;

    const body = document.getElementById('bodyEditor').value;


    let backupType = 'nobackup';
    let backupConfig = '';

    const backupCheckbox = document.getElementById('enableBackup');
    if (backupCheckbox.checked) {
        backupType = 'http'
        backupConfig = {
            'LoadType'  : document.getElementById('LoadType').value,
            'LoadURL'   : document.getElementById('LoadEndpoint').value,

            'RestoreType'  : document.getElementById('RestoreType').value,
            'RestoreURL'   : document.getElementById('RestoreEndpoint').value,

            'Parameters': {
                'load_headers': document.getElementById('loaderHeadersEditor').value,
                'load_body'   : document.getElementById('loaderBodyEditor').value,
                'restore_headers': document.getElementById('restoreHeadersEditor').value,
                'restore_body'   : document.getElementById('restoreBodyEditor').value,
            },
        }
    }

    var reqBody = {
        'Name': name,
        'PatchEnvelopes': patchEnvelopes,
        'Templates': {
            'headers': headers,
            'body': body,
        },
        'DriverType': 'http',
        'DriverConfig': {
            'ReqType': 'GET',
            'OutputURL': endpoint,
        },
        'BackupType': backupType,
        'BackupConfig': backupConfig,
    }
    
    try {
        const response = await fetch('/api/recaster', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(reqBody), 
        });

        const result = await response.text();
        alert(result)

    } catch (error) {
        alert("Error sending or receiving data: " + error.message);
    }
}

function fetchData() {
    fetch("/api/envelopes")
        .then(response => response.json())
        .then(data => {
            if (data.Envelopes && Array.isArray(data.Envelopes)) {
                const checkboxContainer = document.getElementById('envelopesContainer');
                data.Envelopes.forEach(envelope => {
                    const checkbox = document.createElement('input')
                    checkbox.type = 'checkbox';
                    checkbox.id = envelope.PatchID;
                    checkbox.value = envelope.PatchID;

                    const label = document.createElement('label');
                    label.appendChild(document.createTextNode(envelope.PatchID));

                    checkboxContainer.appendChild(checkbox);
                    checkboxContainer.appendChild(label);
                });
            }
        })

    const url = window.location.href;
    const params = new URLSearchParams(new URL(url).search);
    const name = params.get('name');
    const reqUrl = `/api/recasters?name=${encodeURIComponent(name)}`;
    fetch(reqUrl)
        .then(response => response.json())
        .then(data => {
            document.getElementById('name').value = data.Name

            const dropdownContent = document.getElementById('envelopesContainer')

            if (data.PatchEnvelopes && Array.isArray(data.PatchEnvelopes)) {
                data.PatchEnvelopes.forEach(envelope => {
                    console.log(envelope)
                    document.getElementById(envelope).checked = true;
                });
            }

            document.getElementById('headersEditor').value = data.Templates.headers;
            document.getElementById('bodyEditor').value = data.Templates.body;
            
            if (data.BackupType == 'http') {
                document.getElementById('enableBackup').checked = true;
                document.getElementById('backup-container').style.display = 'block';
                
                document.getElementById('LoadType').value = data.BackupConfig.LoadType;
                document.getElementById('LoadEndpoint').value = data.BackupConfig.LoadURL;

                document.getElementById('RestoreType').value = data.BackupConfig.RestoreType;
                document.getElementById('RestoreEndpoint').value = data.BackupConfig.RestoreURL;

                const params = data.BackupConfig.Parameters
                document.getElementById('loaderHeadersEditor').value = params.load_headers;
                document.getElementById('loaderBodyEditor').value = params.load_body; 
                document.getElementById('restoreHeadersEditor').value = params.restore_headers;
                document.getElementById('restoreBodyEditor').value = params.restore_body;
            }

            if (data.DriverType == 'http') {
                document.getElementById('endpoint').value = data.DriverConfig.OutputURL
            }
        })
}

var checkbox = document.getElementById('enableBackup');
var hiddenDiv = document.getElementById('backup-container');
checkbox.addEventListener('change', function() {
  hiddenDiv.style.display = checkbox.checked ? 'block' : 'none';
});

window.onload = fetchData;

var headerReader = document.getElementById("headersEditorButton")
headerReader.addEventListener("click", function() {
    renderJSON('headersEditor');
});

var bodyReader = document.getElementById("bodyEditorButton")
bodyReader.addEventListener("click", function() {
    renderJSON('bodyEditor');
});

var submitForm = document.getElementById("submitForm")
submitForm.addEventListener("click", function() {
    submit();
});

