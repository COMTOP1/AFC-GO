package com.bswdi.utils;

import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.security.spec.InvalidKeySpecException;
import java.sql.Connection;
import java.time.LocalDate;
import java.util.*;

import javax.crypto.SecretKeyFactory;
import javax.crypto.spec.PBEKeySpec;
import javax.servlet.ServletRequest;
import javax.servlet.http.Cookie;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import com.auth0.jwt.*;
import com.auth0.jwt.algorithms.Algorithm;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.bswdi.beans.JWTToken;
import com.bswdi.beans.Users;
import com.bswdi.connection.ConnectionUtils;
import com.bswdi.dotenv.Dotenv;

/**
 * Utilities for java and user data
 *
 * @author BSWDI
 * @version 1.0
 */
public class MyUtils {

    /**
     * Attribute for connection
     */
    private static final String ATT_NAME_CONNECTION = "Q09OTkVDVElPTl9BVFRSSUJVVEU";    // Base64 encoded
    private static final String ATT_JWT_TOKEN = "SldUX1RPS0VOX0FUVFJJQlVURQ";           // Base64 encoded
    private static final Random secureRandom = new SecureRandom();

    /**
     * Sets connection in attributes
     *
     * @param request request
     * @param con     connection
     */
    public static void storeConnection(ServletRequest request, Connection con) {
        request.setAttribute(ATT_NAME_CONNECTION, con);
    }

    /**
     * Return connection from attributes
     *
     * @param request request
     * @return Connection con
     */
    public static Connection getStoredConnection(ServletRequest request) {
        return (Connection) request.getAttribute(ATT_NAME_CONNECTION);
    }

    /**
     * Sets user in cookies
     *
     * @param response response
     * @param user     user
     */
    public static void storeUser(HttpServletRequest request, HttpServletResponse response, Connection con, Users user) {
        try {
            if (con == null)
                con = ConnectionUtils.getConnection();
        } catch (Exception e) {
            e.printStackTrace();
        }
        Dotenv dotenv = Dotenv.load();
        String sessionID = null;
        for (Cookie findSession : request.getCookies())
            if ("JSESSIONID".equals(findSession.getName()))
                sessionID = findSession.getValue();
        JWTToken jwtToken = new JWTToken(0, user.getEmail(), getTime(), getTime1Day(), request.getHeader("user-agent"), sessionID);
        try {
            try {
                assert con != null;
                jwtToken.setId(DBUtils.insertJWTToken(con, jwtToken));
            } catch (Exception e) {
                e.printStackTrace();
            }
            Algorithm algorithm = Algorithm.HMAC512(dotenv.get("JWT_SECRET"));
            String token = JWT.create()
                    .withIssuer("BSWDI")
                    .withAudience("AFC")
                    .withExpiresAt(new Date(jwtToken.getExp()))
                    .withIssuedAt(new Date(jwtToken.getIat()))
                    .withKeyId(String.valueOf(jwtToken.getId()))
                    .sign(algorithm);
            Cookie cookieJWT = new Cookie(ATT_JWT_TOKEN, Base64.getEncoder().encodeToString(token.getBytes()));
            cookieJWT.setMaxAge((int) (jwtToken.getExp() - jwtToken.getIat()));
            cookieJWT.setHttpOnly(true);
            response.addCookie(cookieJWT);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    /**
     * Gets user from cookie
     * This checks that there's a JWT stored in the cookies and then proceeds to verify it
     * The verification will check the user agent (the client you're using) and the session id that tomcat server assigns
     *
     * @param request request
     * @param response response
     * @param con connection
     * @return Users user
     */
    public static Users getUser(HttpServletRequest request, HttpServletResponse response, Connection con) {
        try {
            if (con == null)
                con = ConnectionUtils.getConnection();
        } catch (Exception e) {
            e.printStackTrace();
        }
        Dotenv dotenv = Dotenv.load();
        Algorithm algorithm = Algorithm.HMAC512(dotenv.get("JWT_SECRET"));
        Cookie[] cookies = request.getCookies();
        if (cookies != null)
            for (Cookie cookie : cookies)
                if (ATT_JWT_TOKEN.equals(cookie.getName())) {
                    JWTToken jwtToken = null;
                    try {
                        String token = new String(Base64.getDecoder().decode(cookie.getValue()));
                        JWTVerifier verifier = JWT.require(algorithm)
                                .withIssuer("BSWDI")
                                .withAudience("AFC")
                                .build();
                        DecodedJWT jwt = verifier.verify(token);
                        try {
                            assert con != null;
                            jwtToken = DBUtils.findJWTToken(con, Long.parseLong(jwt.getHeaderClaim("kid").asString()));
                        } catch (Exception e) {
                            deleteUserCookie(response);
                            e.printStackTrace();
                            throw new Exception("Find JWT error");
                        }
                        assert jwtToken != null;
                        String sessionID = null;
                        for (Cookie findSession : cookies)
                            if ("JSESSIONID".equals(findSession.getName()))
                                sessionID = findSession.getValue();
                        if (!jwtToken.validate(request.getHeader("user-agent"), sessionID)) {
                            deleteUserCookie(response);
                            DBUtils.deleteJWTToken(con, jwtToken.getId());
                            return null;
                        }
                        return DBUtils.findUser(con, jwtToken.getEmail());
                    } catch (Exception e) {
                        if (jwtToken != null)
                            try {
                                DBUtils.deleteJWTToken(con, jwtToken.getId());
                            } catch (Exception f) {
                                f.printStackTrace();
                            }
                        deleteUserCookie(response);
                        String[] fullException = e.getClass().getCanonicalName().split("\\.");
                        String exception = fullException[fullException.length - 1];
                        String error = null;
                        try {
                            error = e.toString().split("\n")[0].split(":")[1].trim();
                        } catch (Exception f) {
                            e.printStackTrace();
                        }
                        System.out.printf("%s with error message \"%s\"\nINVALID", exception, error);
                        return null;
                    }
                }
        return null;
    }

    /**
     * Delete all cookies
     *
     * @param request  request
     * @param response response
     */
    public static void deleteAllCookies(HttpServletRequest request, HttpServletResponse response) {
        Cookie[] cookies = request.getCookies();
        for (Cookie cookie : cookies) {
            cookie.setMaxAge(0);
            response.addCookie(cookie);
        }
    }

    /**
     * Delete user from cookies
     *
     * @param response response
     */
    public static void deleteUserCookie(HttpServletResponse response) {
        Cookie cookieJWTToken = new Cookie(ATT_JWT_TOKEN, null);
        cookieJWTToken.setMaxAge(0);
        response.addCookie(cookieJWTToken);
    }

    public static long getTime1Day() {
        Date date = new Date();
        Calendar cal = Calendar.getInstance();
        cal.setTime(date);
        cal.add(Calendar.DATE, 1);
        return cal.getTime().getTime();
    }

    /**
     * Return time
     *
     * @return String time
     */
    public static long getTime() {
        Date date = new Date();
        return date.getTime();
    }

    /**
     * Return year
     *
     * @return int year
     */
    @SuppressWarnings("deprecation")
    public static int getYear() {
        Date date = new Date();
        return date.getYear() + 1900;
    }

    /**
     * Return epoch
     *
     * @return long epoch
     */
    public static long getEpoch() {
        LocalDate date = LocalDate.now();
        return date.toEpochDay();
    }

    /**
     * Returns a random salt to be used to hash a password.
     *
     * @return a random salt with specified length
     */
    public static byte[] getNextSalt() {
        Dotenv dotenv = Dotenv.load();
        int keyLengthBytes = Integer.parseInt(dotenv.get("KEY_LENGTH_BYTES"));
        byte[] salt = new byte[keyLengthBytes];
        secureRandom.nextBytes(salt);
        return salt;
    }

    /**
     * Returns a salted and hashed password using the provided hash.<br>
     * Note - side effect: the password is destroyed (the char[] is filled with zeros)
     *
     * @param password the password to be hashed
     * @param salt     a 64 bytes salt, ideally obtained with the getNextSalt method
     * @return the hashed password with a pinch of salt
     */
    public static byte[] hash(char[] password, byte[] salt) {
        Dotenv dotenv = Dotenv.load();
        int iterations = Integer.parseInt(dotenv.get("ITERATIONS")), keyLengthBits = Integer.parseInt(dotenv.get("KEY_LENGTH_BITS"));
        PBEKeySpec spec = new PBEKeySpec(password, salt, iterations, keyLengthBits);
        Arrays.fill(password, Character.MIN_VALUE);
        try {
            return SecretKeyFactory.getInstance("PBKDF2WithHmacSHA512").generateSecret(spec).getEncoded();
        } catch (NoSuchAlgorithmException | InvalidKeySpecException e) {
            throw new AssertionError("Error while hashing a password: " + e.getMessage(), e);
        } finally {
            spec.clearPassword();
        }
    }

    /**
     * Returns true if the given password and salt match the hashed value, false otherwise.<br>
     * Note - side effect: the password is destroyed (the char[] is filled with zeros)
     *
     * @param password     the password to check
     * @param salt         the salt used to hash the password
     * @param expectedHash the expected hashed value of the password
     * @return true if the given password and salt match the hashed value, false otherwise
     */
    public static boolean verifyPassword(char[] password, byte[] salt, byte[] expectedHash) {
        byte[] pwdHash = hash(password, salt);
        Arrays.fill(password, Character.MIN_VALUE);
        if (pwdHash.length != expectedHash.length) return false;
        for (int i = 0; i < pwdHash.length; i++)
            if (pwdHash[i] != expectedHash[i]) return false;
        return true;
    }
}
