function Update(mesh) {
    var R = document.getElementById('R').value;
    var r = document.getElementById('r').value;
    var phi = document.getElementById('phi').value;
    var thick = document.getElementById('t').value;

    url = "/png?f=png";
    url = url.concat("&R=", R);
    url = url.concat("&r=", r);
    url = url.concat("&phi=", phi);
    url = url.concat("&t=", thick);
  
    document.getElementById("pattern").src=url;

    // Now the GL stuff:
    var helper  = function(u, v, r, R, phi_deg) {
	var theta = v * 2 * Math.PI;
	var x_disp = r*Math.sin(theta)
	var d = 4
	if (Math.abs(x_disp) < R) {
	    d -= Math.sqrt(R*R - x_disp*x_disp)
	}

	if (phi_deg != 90) {
	    var phi = phi_deg / 180 * Math.PI;
	    d += (r - r*Math.cos(theta)) / Math.tan(phi)
	}

	if (d > 4) d = 4;
	
	var x = r*Math.sin(theta)
	var y = r*Math.cos(theta)
	var z = u * d
	return new THREE.Vector3(x, y, z);
    }

    var paramfunc = function (u, v) {
	if (u < 0.5) {
	    return helper(2*u, v, r - thick, R, phi)
	} else {
	    return helper(2*(1-u), v, r, R, phi)
	}
    }

    mesh.children[0].geometry.dispose();
    mesh.children[0].geometry = new THREE.ParametricBufferGeometry( paramfunc, 50, 50 );
}

function Setup() {
    var scene = new THREE.Scene();
    var camera = new THREE.OrthographicCamera(-4, 4, 4, -4,  0.1, 2000 );

    var renderer = new THREE.WebGLRenderer();
    renderer.setSize( 800, 800 );
    document.body.appendChild( renderer.domElement );
    
    var lights = [];
    lights[ 0 ] = new THREE.PointLight( 0xffffff, 1, 0 );
    lights[ 1 ] = new THREE.PointLight( 0xffffff, 1, 0 );
    lights[ 2 ] = new THREE.PointLight( 0xffffff, 1, 0 );

    lights[ 0 ].position.set( 0, 200, 0 );
    lights[ 1 ].position.set( 100, 200, 100 );
    lights[ 2 ].position.set( - 100, - 200, - 100 );

    scene.add( lights[ 0 ] );
    scene.add( lights[ 1 ] );
    scene.add( lights[ 2 ] );

    var mesh = new THREE.Object3D();
    mesh.add ( new THREE.Mesh(
	new THREE.Geometry(), // do be replaced in Update()
	new THREE.MeshStandardMaterial( {
	    color: 0xA0A0A0,
	    side: THREE.DoubleSide,
	    roughness: 0.74,
	    metalness: 1.0,
	    flatShading: true,
	} )));

    // So mesh is an object and apparently JS will bind a mutable reference into the closure?
    document.getElementById("preview").addEventListener("click", function() {Update(mesh)});

    scene.add(mesh);
    camera.position.z = 5;
    camera.position.x = 5;
    camera.up.set(0,0,1);
    camera.lookAt(new THREE.Vector3(0,0,2));

    var animate = function () {
	requestAnimationFrame( animate );
	mesh.rotation.z += 0.005
	renderer.render(scene, camera);
    };

    animate();
    Update(mesh);
}
